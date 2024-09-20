/*
Package cache provides a simple in-memory cache implementation.
*/
package cache

import (
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"syscall"

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
	sizeInBytes    int
	evictionPolicy EvictionPolicy
	maxSizeInBytes int
	evictPathCh    chan string
	mu             sync.RWMutex
	synched        bool
}

/*
ZKCacheOptions is used to configure the cache.
*/
type ZKCacheOptions struct {
	// MaxSizeInBytes is the maximum size of the cache in bytes.
	MaxSizeInBytes int
	// EvictionPolicy is the policy used to evict nodes from the cache.
	EvictionPolicy EvictionPolicy
	// EnableCacheSynch is a flag to enable cache synchronization with the ZooKeeper server on node data change.
	EnableCacheSynch bool
}

/*
ZKCacheOptionsBuilder is a builder for ZKCacheOptions.
*/
type ZKCacheOptionsBuilder struct {
	maxSizeInBytes   int
	evictionPolicy   EvictionPolicy
	enableCacheSynch bool
}

const (
	defaultCacheMemoryPercentage = 5
)

/*
NewCacheOptionsBuilder creates a new ZKCacheOptionsBuilder.
*/
func NewCacheOptionsBuilder() (ZKCacheOptionsBuilder, error) {
	var sysinfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysinfo)
	if err != nil {
		return ZKCacheOptionsBuilder{}, err
	}
	availableMemory := sysinfo.Totalram * uint64(sysinfo.Unit)

	pctg, ok := os.LookupEnv("ZK_CACHE_MAX_SIZE_PCTG")
	useCachePctg := defaultCacheMemoryPercentage

	if ok {
		parsedPctg, err := strconv.Atoi(pctg)
		if err == nil && parsedPctg >= 0 && parsedPctg <= 100 {
			useCachePctg = parsedPctg
		} else {
			log.Printf("Invalid value for ZK_CACHE_MAX_SIZE_PCTG: %s. Using default value of %d", pctg, useCachePctg)
		}
	}
	maxSizeInBytes := int(availableMemory * uint64(useCachePctg) / 100)

	return ZKCacheOptionsBuilder{
		maxSizeInBytes:   maxSizeInBytes,
		evictionPolicy:   EvictLeastRecentlyUsed,
		enableCacheSynch: true,
	}, nil
}

/*
WithMaxSizeInBytes sets the maximum size of the cache in bytes.
*/
func (b ZKCacheOptionsBuilder) WithMaxSizeInBytes(maxSizeInBytes int) ZKCacheOptionsBuilder {
	b.maxSizeInBytes = maxSizeInBytes
	return b
}

/*
WithEvictionPolicy sets the eviction policy for the cache.
*/
func (b ZKCacheOptionsBuilder) WithEvictionPolicy(evictionPolicy EvictionPolicy) ZKCacheOptionsBuilder {
	b.evictionPolicy = evictionPolicy
	return b
}

/*
WithEnableCacheSynch sets the flag to enable cache synchronization with the ZooKeeper server on node data change.
*/
func (b ZKCacheOptionsBuilder) WithEnableCacheSynch(enableCacheSynch bool) ZKCacheOptionsBuilder {
	b.enableCacheSynch = enableCacheSynch
	return b
}

/*
Build builds the ZKCacheOptions.
*/
func (b ZKCacheOptionsBuilder) Build() ZKCacheOptions {
	return ZKCacheOptions{
		MaxSizeInBytes:   b.maxSizeInBytes,
		EvictionPolicy:   b.evictionPolicy,
		EnableCacheSynch: b.enableCacheSynch,
	}
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
func (c *Cache) Get(nodeName string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	actualPath := path.Join(append([]string{c.framework.Namespace()}, nodeName)...)

	cachedData, ok := c.cache[actualPath]
	if ok {
		return cachedData, nil
	}

	data, err := operation.Get(c.framework, actualPath)
	if err != nil {
		return nil, err
	}
	c.cache[actualPath] = data
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
