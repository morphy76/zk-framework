/*
Package cache provides a simple in-memory cache implementation.
*/
package cache

import (
	"log"
	"math"
	"path"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/cache/cacheerr"
	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/operation"
	"github.com/morphy76/zk/pkg/watcher"
)

/*
EvictionPolicy is the policy used to evict nodes from the cache.
*/
type EvictionPolicy int

const (
	// EvictLeastRecentlyUsed evicts the least recently used node.
	EvictLeastRecentlyUsed EvictionPolicy = iota
	// EvictLeastFrequentlyUsed evicts the least frequently used node.
	EvictLeastFrequentlyUsed
	// EvictRandomly evicts a random node.
	EvictRandomly
)

/*
Cache is a simple in-memory cache implementation.
*/
type Cache struct {
	framework      core.ZKFramework
	cache          map[string][]byte
	cacheUsage     map[string]int64
	sizeInBytes    int
	evictionPolicy EvictionPolicy
	maxSizeInBytes int
	evictPathCh    chan string
	mu             sync.RWMutex
	synched        bool
}

/*
NewCache creates a new cache using the default eviction policy, which is EvictLeastRecentlyUsed.
*/
func NewCache(framework core.ZKFramework) (*Cache, error) {

	builder, err := NewCacheOptionsBuilder()
	if err != nil {
		return nil, err
	}

	return NewCacheWithOptions(framework, builder.Build())
}

/*
NewCacheWithOptions creates a new cache specifying the cache options.
*/
func NewCacheWithOptions(framework core.ZKFramework, options ZKCacheOptions) (*Cache, error) {

	if options.MaxSizeInBytes <= 0 {
		return nil, cacheerr.ErrInvalidCacheSize
	}

	return &Cache{
		framework:      framework,
		cache:          make(map[string][]byte),
		cacheUsage:     make(map[string]int64),
		sizeInBytes:    0,
		evictionPolicy: options.EvictionPolicy,
		maxSizeInBytes: options.MaxSizeInBytes,
		synched:        options.EnableCacheSynch,
		evictPathCh:    make(chan string),
		mu:             sync.RWMutex{},
	}, nil
}

/*
Clear clears the cache.
*/
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for zkPath := range c.cache {
		c.evict(zkPath)
	}
	c.refreshSizeInBytes()
}

/*
Get gets a node at the given path.
*/
func (c *Cache) IsCached(nodeName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	actualPath := path.Join(append([]string{c.framework.Namespace()}, nodeName)...)

	_, ok := c.cache[actualPath]
	return ok
}

/*
Get gets a node at the given path.
*/
func (c *Cache) Get(nodeName string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	actualPath := path.Join(append([]string{c.framework.Namespace()}, nodeName)...)

	cachedData, ok := c.cache[actualPath]
	if ok {
		c.incrementUsageByPolicy(actualPath)
		return cachedData, nil
	}

	if c.testExceedingResources() {
		err := c.evictByPolicy()
		if err != nil {
			log.Printf("Error evicting cache: %v, warning, possible leak", err)
		}
	}

	data, err := operation.Get(c.framework, actualPath)
	if err != nil {
		return nil, err
	}
	c.cache[actualPath] = data
	c.initCacheUsageByPolicy(actualPath)
	c.refreshSizeInBytes()

	if !c.synched {
		return data, nil
	}

	outChan := make(chan zk.Event)
	watcher.Set(c.framework, nodeName, outChan, zk.EventNodeDataChanged)
	go func() {
		for {
			select {
			case evictedPath := <-c.evictPathCh:
				if evictedPath == actualPath {
					watcher.UnSet(c.framework, nodeName, zk.EventNodeDataChanged)
					close(outChan)
					return
				}
			case <-outChan:
				c.renew(actualPath)
			}
		}
	}()

	return data, nil
}

/*
GetSizeInBytes returns the size of the cache in bytes.
*/
func (c *Cache) GetSizeInBytes() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.sizeInBytes
}

func (c *Cache) refreshSizeInBytes() {
	size := 0
	for _, data := range c.cache {
		size += len(data)
	}
	c.sizeInBytes = size
}

func (c *Cache) evict(zkPath string) {
	if c.synched {
		c.evictPathCh <- zkPath
	}
	delete(c.cache, zkPath)
	delete(c.cacheUsage, zkPath)
}

func (c *Cache) renew(actualPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := operation.Get(c.framework, actualPath)
	if err != nil {
		log.Printf("Error renewing cache for path %s: %v", actualPath, err)
		delete(c.cache, actualPath)
	}
	c.cache[actualPath] = data
	c.refreshSizeInBytes()

	return nil
}

func (c *Cache) testExceedingResources() bool {
	log.Printf("Cache size: %d, max size: %d", c.sizeInBytes, c.maxSizeInBytes)
	return c.sizeInBytes > c.maxSizeInBytes
}

func (c *Cache) evictByPolicy() error {
	switch c.evictionPolicy {
	case EvictLeastRecentlyUsed:
		return c.evictLRU()
	case EvictLeastFrequentlyUsed:
		return c.evictLFU()
	case EvictRandomly:
		return c.evictRandomly()
	default:
		return cacheerr.ErrInvalidEvictionPolicy
	}
}

func (c *Cache) initCacheUsageByPolicy(zkPath string) {
	if c.evictionPolicy == EvictLeastFrequentlyUsed {
		c.cacheUsage[zkPath] = 1
	} else if c.evictionPolicy == EvictLeastRecentlyUsed {
		c.cacheUsage[zkPath] = time.Now().UnixNano()
	}
}

func (c *Cache) incrementUsageByPolicy(zkPath string) {
	if c.evictionPolicy == EvictLeastFrequentlyUsed {
		c.cacheUsage[zkPath]++
	} else if c.evictionPolicy == EvictLeastRecentlyUsed {
		c.cacheUsage[zkPath] = time.Now().UnixNano()
	}
}

func (c *Cache) evictLRU() error {
	oldestPath := ""
	oldestTime := time.Now().UnixNano()
	for zkPath, time := range c.cacheUsage {
		if time < oldestTime {
			oldestTime = time
			oldestPath = zkPath
		}
	}
	log.Printf("Evicting LRU: %s", oldestPath)
	if oldestPath != "" {
		c.evict(oldestPath)
	}
	return nil
}

func (c *Cache) evictLFU() error {
	leastFrequentPath := ""
	var leastFrequency int64 = math.MaxInt64
	for zkPath, frequency := range c.cacheUsage {
		if frequency < leastFrequency {
			leastFrequency = frequency
			leastFrequentPath = zkPath
		}
	}
	log.Printf("Evicting LFU: %s", leastFrequentPath)
	if leastFrequentPath != "" {
		c.evict(leastFrequentPath)
	}
	return nil
}

func (c *Cache) evictRandomly() error {
	log.Printf("Evicting randomly")
	for zkPath := range c.cache {
		c.evict(zkPath)
		break
	}
	return nil

}
