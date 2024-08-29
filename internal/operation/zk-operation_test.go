package operation_test

import (
	"os"
	"testing"
	"time"

	"github.com/morphy76/zk/internal/framework"
	"github.com/morphy76/zk/internal/operation"
	testutil "github.com/morphy76/zk/internal/test_util"
)

const (
	zkHostEnv = "ZK_HOST"
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

func TestZKOperation(t *testing.T) {

	t.Run("List nodes", func(t *testing.T) {
		t.Log("List nodes")
		zkFramework, err := framework.CreateFramework("")
		if !framework.IsInvalidConnectionURL(err) {
			t.Errorf("expected error %v, got %v", framework.ErrInvalidConnectionURL, err)
		}

		err = zkFramework.WaitConnection(5 * time.Second)
		if !framework.IsConnectionTimeout(err) {
			t.Errorf("expected error %v, got %v", framework.ErrConnectionTimeout, err)
		}

		if !zkFramework.Connected() {
			t.Errorf("expected client to be connected")
		}

		operation.Ls(zkFramework)
	})
}
