/*
Package cache provides a simple in-memory cache implementation.
*/
package cache

import (
	"log"
	"os"
	"strconv"
	"syscall"
)

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
