package operation_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/morphy76/zk/pkg/framework"
	"github.com/morphy76/zk/pkg/operation"
	testutil "github.com/morphy76/zk/pkg/test_util"
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
			t.Error(expectedClientToBeConnected)
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

	t.Run("Create node", func(t *testing.T) {
		t.Log("Create node")
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
			t.Error(expectedClientToBeConnected)
		}

		nodeName := uuid.New().String()

		if err := operation.Create(zkFramework, nodeName); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("Exists node", func(t *testing.T) {
		t.Log("Exists node")
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
			t.Error(expectedClientToBeConnected)
		}

		nodeName := path.Join(uuid.New().String(), uuid.New().String())
		if err := operation.Create(zkFramework, nodeName); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		exists, err := operation.Exists(zkFramework, nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if !exists {
			t.Errorf("expected node to exist")
		}
	})
}
