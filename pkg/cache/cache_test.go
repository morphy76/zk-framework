package cache_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/internal/test_util/mocks"
	"github.com/morphy76/zk/pkg/cache"
	"github.com/morphy76/zk/pkg/cache/cacheerr"
	"github.com/morphy76/zk/pkg/operation"
)

const (
	zkHostEnv                   = "ZK_HOST"
	unexpectedErrorFmt          = "unexpected error %v"
	expectedClientToBeConnected = "expected client to be connected"
)

func TestMain(m *testing.M) {
	zkC, ctx, err := testutil.StartTestServer()
	if err != nil {
		panic(err)
	}
	defer zkC.Terminate(ctx)

	host, err := zkC.Host(ctx)
	if err != nil {
		panic(err)
	}
	mappedPort, err := zkC.MappedPort(ctx, "2181")
	if err != nil {
		panic(err)
	}
	os.Setenv(zkHostEnv, host+":"+mappedPort.Port())

	exitCode := m.Run()

	os.Unsetenv(zkHostEnv)
	os.Exit(exitCode)
}

func TestZKCache(t *testing.T) {

	t.Run("Create the cache with default options", func(t *testing.T) {
		t.Log("Create the cache with default options")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		zkCache, err := cache.NewCache(zkFramework)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkCache.Clear()
	})

	t.Run("Create the cache with bad options, negative max cache size", func(t *testing.T) {
		t.Log("Create the cache with bad options, negative max cache size")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		optsBuilder, err := cache.NewCacheOptionsBuilder()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		opts := optsBuilder.WithMaxSizeInBytes(-1).Build()

		_, err = cache.NewCacheWithOptions(zkFramework, opts)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		} else if !cacheerr.IsInvalidCacheSize(err) {
			t.Fatalf("Expected invalid cache size error, got %v", err)
		}
	})

	t.Run("Initial get data from the cache", func(t *testing.T) {
		t.Log("Initial get data from the cache")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		spiedFramework := mocks.NewSpiedFramework(zkFramework)

		zkCache, err := cache.NewCache(spiedFramework)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkCache.Clear()

		nodeName := uuid.New().String()
		data := []byte(uuid.New().String())

		opts := operation.NewCreateOptionsBuilder().
			WithData(data).
			Build()

		if err := operation.CreateWithOptions(zkFramework, nodeName, opts); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if zkCache.GetSizeInBytes() != 0 {
			t.Errorf("Expected cache size to be 0, got %v", zkCache.GetSizeInBytes())
		}
		cachedData, err := zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if string(cachedData) != string(data) {
			t.Errorf("Expected data to be %v, got %v", data, cachedData)
		}

		if spiedFramework.Interactions["Cn"] != 2 {
			t.Errorf("Expected Cn to be called once but was called %v times", spiedFramework.Interactions["Cn"])
		}

		if zkCache.GetSizeInBytes() != len(data) {
			t.Errorf("Expected cache size to be %v, got %v", len(data), zkCache.GetSizeInBytes())
		}
	})

	t.Run("Get data from the cache", func(t *testing.T) {
		t.Log("Get data from the cache")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		spiedFramework := mocks.NewSpiedFramework(zkFramework)

		zkCache, err := cache.NewCache(spiedFramework)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkCache.Clear()

		nodeName := uuid.New().String()
		data := []byte(uuid.New().String())

		opts := operation.NewCreateOptionsBuilder().
			WithData(data).
			Build()

		if err := operation.CreateWithOptions(zkFramework, nodeName, opts); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if zkCache.GetSizeInBytes() != 0 {
			t.Errorf("Expected cache size to be 0, got %v", zkCache.GetSizeInBytes())
		}
		_, err = zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkCache.GetSizeInBytes() != len(data) {
			t.Errorf("Expected cache size to be %v, got %v", len(data), zkCache.GetSizeInBytes())
		}

		cachedData, err := zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if string(cachedData) != string(data) {
			t.Errorf("Expected data to be %v, got %v", data, cachedData)
		}

		if spiedFramework.Interactions["Cn"] != 2 {
			t.Errorf("Expected Cn to be called once but was called %v times", spiedFramework.Interactions["Cn"])
		}

		if zkCache.GetSizeInBytes() != len(data) {
			t.Errorf("Expected cache size to be %v, got %v", len(data), zkCache.GetSizeInBytes())
		}
	})

	t.Run("Get data from a synched cache", func(t *testing.T) {
		t.Log("Get data from a synched cache")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		spiedFramework := mocks.NewSpiedFramework(zkFramework)

		zkCache, err := cache.NewCache(spiedFramework)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkCache.Clear()

		nodeName := uuid.New().String()
		data := []byte(uuid.New().String())

		opts := operation.NewCreateOptionsBuilder().
			WithData(data).
			Build()

		if err := operation.CreateWithOptions(zkFramework, nodeName, opts); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		_, err = zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		cachedData, err := zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if string(cachedData) != string(data) {
			t.Errorf("Expected data to be %v, got %v", data, cachedData)
		}

		newData := []byte(uuid.New().String())
		operation.Update(zkFramework, nodeName, newData)
		<-time.After(1 * time.Second)

		cachedData, err = zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if string(cachedData) == string(data) {
			t.Errorf("Expected data to be updated")
		}

		if string(cachedData) != string(newData) {
			t.Errorf("Expected data to be %v, got %v", string(cachedData), string(newData))
		}
	})

	t.Run("Get data from a non-synched cache", func(t *testing.T) {
		t.Log("Get data from a non-synched cache")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		spiedFramework := mocks.NewSpiedFramework(zkFramework)

		optsBuilder, err := cache.NewCacheOptionsBuilder()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		cacheOpts := optsBuilder.WithEnableCacheSynch(false).Build()

		zkCache, err := cache.NewCacheWithOptions(spiedFramework, cacheOpts)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkCache.Clear()

		nodeName := uuid.New().String()
		data := []byte(uuid.New().String())

		opts := operation.NewCreateOptionsBuilder().
			WithData(data).
			Build()

		if err := operation.CreateWithOptions(zkFramework, nodeName, opts); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		_, err = zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		cachedData, err := zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if string(cachedData) != string(data) {
			t.Errorf("Expected data to be %v, got %v", data, cachedData)
		}

		newData := []byte(uuid.New().String())
		operation.Update(zkFramework, nodeName, newData)
		<-time.After(1 * time.Second)

		cachedData, err = zkCache.Get(nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if string(cachedData) != string(data) {
			t.Errorf("Expected data to be not updated because it is not in sync")
		}
	})
}
