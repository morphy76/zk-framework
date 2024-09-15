package cache_test

import (
	"os"
	"testing"

	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/pkg/framework"
	"github.com/morphy76/zk/pkg/framework/frwkerr"
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
		_, err := framework.CreateFramework("")
		if !frwkerr.IsInvalidConnectionURL(err) {
			t.Errorf("expected error %v, got %v", frwkerr.ErrInvalidConnectionURL, err)
		}
	})
}
