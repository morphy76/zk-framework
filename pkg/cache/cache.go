/*
Package cache provides a simple in-memory cache implementation.
*/
package cache

import (
	"log"
	"os"
	"strconv"
	"syscall"

	"github.com/morphy76/zk/pkg/cache/cacheerr"
	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/operation"
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
}

/*
ZKCacheOptions is used to configure the cache.
*/
type ZKCacheOptions struct {
	MaxSizeInBytes int
	EvictionPolicy EvictionPolicy
}

const (
	defaultCacheMemoryPercentage = 5
)

/*
NewCache creates a new cache using the default eviction policy, which is EvictLeastRecentlyUsed.
*/
func NewCache(framework core.ZKFramework) (*Cache, error) {

	var sysinfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysinfo)
	if err != nil {
		return nil, err
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

	return NewCacheWithOptions(framework, ZKCacheOptions{
		EvictionPolicy: EvictLeastRecentlyUsed,
		MaxSizeInBytes: maxSizeInBytes,
	})
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
	}, nil
}

/*
Clear clears the cache.
*/
func (c *Cache) Clear() {
	c.cache = make(map[string][]byte)
	c.refreshSizeInBytes()
}

/*
Get gets a node at the given path.
*/
func (c *Cache) Get(nodeName string) ([]byte, error) {

	cachedData, ok := c.cache[nodeName]
	if ok {
		return cachedData, nil
	}

	data, err := operation.Get(c.framework, nodeName)
	if err != nil {
		return nil, err
	}
	c.cache[nodeName] = data
	c.refreshSizeInBytes()
	return data, nil
}

/*
GetSizeInBytes returns the size of the cache in bytes.
*/
func (c *Cache) GetSizeInBytes() int {
	return c.sizeInBytes
}

func (c *Cache) refreshSizeInBytes() {
	size := 0
	for _, data := range c.cache {
		size += len(data)
	}
	c.sizeInBytes = size
}
