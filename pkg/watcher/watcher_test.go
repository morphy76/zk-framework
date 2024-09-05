package watcher_test

import (
	"os"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/pkg/framework"
	"github.com/morphy76/zk/pkg/operation"
	"github.com/morphy76/zk/pkg/watcher"
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

func TestZKWatcher(t *testing.T) {

	t.Run("Monitor and notify node changes", func(t *testing.T) {
		t.Log("Set a watcher")
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

		events := make(chan zk.Event)
		if err := watcher.Set(zkFramework, nodeName, events, zk.EventNodeDataChanged); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		nodeData := []byte(uuid.New().String())
		operation.Update(zkFramework, nodeName, nodeData)

		zkEvent := <-events
		if zkEvent.Type != zk.EventNodeDataChanged {
			t.Errorf("expected %v, got %v", zk.EventNodeDataChanged, zkEvent.Type)
		}
		t.Logf("Received event %v", zkEvent)
	})

	t.Run("monitor a non-existent node", func(t *testing.T) {
		t.Log("Set a watcher")
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
		events := make(chan zk.Event)
		if err := watcher.Set(zkFramework, nodeName, events, zk.EventNodeDataChanged); err != watcher.ErrUnknownNode {
			t.Errorf("expected %v, got %v", watcher.ErrUnknownNode, err)
		}
	})

	t.Run("monitor the same node, twice", func(t *testing.T) {
		t.Log("Set a watcher twice")
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

		events := make(chan zk.Event)
		if err := watcher.Set(zkFramework, nodeName, events, zk.EventNodeDataChanged); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := watcher.Set(zkFramework, nodeName, events, zk.EventNodeDataChanged); err == nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("monitor the same node, different events", func(t *testing.T) {
		t.Log("Set a watcher twice for different events")
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

		events := make(chan zk.Event)
		if err := watcher.Set(zkFramework, nodeName, events, zk.EventNodeDataChanged); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := watcher.Set(zkFramework, nodeName, events, zk.EventNodeChildrenChanged); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})
}
