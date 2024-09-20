package cache_test

import (
	"math/rand"
	"testing"

	"github.com/morphy76/zk/pkg/cache"
)

func TestDefaultCacheOptionsBuilder(t *testing.T) {
	builder, err := cache.NewCacheOptionsBuilder()
	if err != nil {
		t.Errorf(unexpectedErrorFmt, err)
	}
	opts := builder.Build()

	if !opts.EnableCacheSynch {
		t.Errorf("Expected EnableCacheSynch to be true, got false")
	}

	if opts.MaxSizeInBytes == 0 {
		t.Errorf("Expected MaxSizeInBytes to be > 0, got %d", opts.MaxSizeInBytes)
	}

	if opts.EvictionPolicy != cache.EvictLeastRecentlyUsed {
		t.Errorf("Expected EvictionPolicy to be %v, got %v", cache.EvictLeastRecentlyUsed, opts.EvictionPolicy)
	}
}

func TestCacheOptionsBuilder(t *testing.T) {
	evictPolicy := cache.EvictLeastFrequentlyUsed
	sinch := false
	maxSize := rand.Intn(1000) + 1

	builder, err := cache.NewCacheOptionsBuilder()
	if err != nil {
		t.Errorf(unexpectedErrorFmt, err)
	}
	opts := builder.
		WithEvictionPolicy(evictPolicy).
		WithEnableCacheSynch(sinch).
		WithMaxSizeInBytes(maxSize).
		Build()

	if opts.EnableCacheSynch != sinch {
		t.Errorf("Expected EnableCacheSynch to be %v, got %v", sinch, opts.EnableCacheSynch)
	}

	if opts.MaxSizeInBytes != maxSize {
		t.Errorf("Expected MaxSizeInBytes to be %d, got %d", maxSize, opts.MaxSizeInBytes)
	}

	if opts.EvictionPolicy != evictPolicy {
		t.Errorf("Expected EvictionPolicy to be %v, got %v", evictPolicy, opts.EvictionPolicy)
	}
}
