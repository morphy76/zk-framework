package cache_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/internal/test_util/mocks"
	"github.com/morphy76/zk/pkg/cache"
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

	t.Run("Create the cache", func(t *testing.T) {
		t.Log("Create the cache")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		zkCache, err := cache.NewCache(zkFramework)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkCache.Close()
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
		defer zkCache.Close()

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

		if spiedFramework.Interactions["Cn"] != 1 {
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
		defer zkCache.Close()

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

		if spiedFramework.Interactions["Cn"] != 1 {
			t.Errorf("Expected Cn to be called once but was called %v times", spiedFramework.Interactions["Cn"])
		}

		if zkCache.GetSizeInBytes() != len(data) {
			t.Errorf("Expected cache size to be %v, got %v", len(data), zkCache.GetSizeInBytes())
		}
	})
}
