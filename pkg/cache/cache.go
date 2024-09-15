/*
Package cache provides a simple in-memory cache implementation.
*/
package cache

import (
	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/operation"
)

/*
Cache is a simple in-memory cache implementation.
*/
type Cache struct {
	framework   core.ZKFramework
	cache       map[string][]byte
	sizeInBytes int
}

/*
NewCache creates a new cache.
*/
func NewCache(framework core.ZKFramework) (*Cache, error) {
	return &Cache{
		framework: framework,
		cache:     make(map[string][]byte),
	}, nil
}

/*
Close closes the cache.
*/
func (c *Cache) Close() {
	c.cache = make(map[string][]byte)
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
