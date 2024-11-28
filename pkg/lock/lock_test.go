package lock_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/pkg/lock"
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

func TestZKLock(t *testing.T) {
	t.Run("Create a read lock", func(t *testing.T) {
		t.Log("Create a read lock")
		t.Skip("skipping test")
		zkFramework, err := testutil.ConnectFramework()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		subject := uuid.New().String()

		lockable, err := lock.NewLockableBuilder().
			WithSubject(subject).
			Build()
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}

		zkLock := lock.NewLock("test")

		releaseFn, err := zkLock.RAcquire(zkFramework, lockable, 10*time.Second)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		defer releaseFn()

		lockType, err := zkLock.HasLock(zkFramework, lockable)
		if err != nil {
			t.Fatalf(unexpectedErrorFmt, err)
		}
		if lockType != lock.RLock {
			t.Fatalf("expected lock type %v, got %v", lock.RLock, lockType)
		}
	})
}
