package operation_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/uuid"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/pkg/framework"
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

	t.Run("Create a duplicated node", func(t *testing.T) {
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

		if err := operation.Create(zkFramework, nodeName); err == nil {
			t.Error("expected error to be not nil")
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

	t.Run("Delete node", func(t *testing.T) {
		t.Log("Delete node")
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

		if err := operation.Delete(zkFramework, nodeName); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("Delete non-existent node", func(t *testing.T) {
		t.Log("Delete non-existent node")
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
		if err := operation.Delete(zkFramework, nodeName); err == nil {
			t.Error("expected error to be not nil")
		}
	})

	t.Run("Update node", func(t *testing.T) {
		t.Log("Update node")
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

		data := []byte(uuid.New().String())
		version, err := operation.Update(zkFramework, nodeName, []byte(data))
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if version == 0 {
			t.Errorf("expected version to be non-zero")
		}
		t.Logf("Updated node with version: %d", version)

		readData, err := operation.Get(zkFramework, nodeName)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if string(readData) != string(data) {
			t.Errorf("expected data to be %s, got %s", string(data), string(readData))
		}
	})

	t.Run("Update non-existent node", func(t *testing.T) {
		t.Log("Update non-existent node")
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
		data := []byte(uuid.New().String())
		_, err = operation.Update(zkFramework, nodeName, []byte(data))
		if err == nil {
			t.Error("expected error to be not nil")
		}
	})

	t.Run("Get non-existent node", func(t *testing.T) {
		t.Log("Get non-existent node")
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
		_, err = operation.Get(zkFramework, nodeName)
		if err == nil {
			t.Error("expected error to be not nil")
		}
	})
}
