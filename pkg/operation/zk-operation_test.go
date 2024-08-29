package operation_test

import (
	"os"
	"testing"
	"time"

	"github.com/morphy76/zk/pkg/framework"
	"github.com/morphy76/zk/pkg/operation"
	testutil "github.com/morphy76/zk/pkg/test_util"
)

const (
	zkHostEnv          = "ZK_HOST"
	unexpectedErrorFmt = "unexpected error %v"
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
		zkFramework, err := framework.CreateFramework(os.Getenv(zkHostEnv))
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Start(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		err = zkFramework.WaitConnection(10 * time.Second)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if !zkFramework.Connected() {
			t.Errorf("expected client to be connected")
		}

		nodes, err := operation.Ls(zkFramework, "/")
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if len(nodes) == 0 {
			t.Errorf("expected non-zero nodes, got %d", len(nodes))
		}
		for _, node := range nodes {
			t.Log(node)
		}
	})
}
