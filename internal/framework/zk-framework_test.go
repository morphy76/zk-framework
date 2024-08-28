package framework_test

import (
	"os"
	"testing"
	"time"

	"github.com/morphy76/zk/internal/framework"
	testutil "github.com/morphy76/zk/internal/test_util"
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

func TestZKFramework(t *testing.T) {

	t.Run("Create a ZK framework with empty URL", func(t *testing.T) {
		_, err := framework.CreateFramework("")
		if !framework.IsInvalidConnectionURL(err) {
			t.Errorf("expected error %v, got %v", framework.ErrInvalidConnectionURL, err)
		}
	})

	t.Run("Create a non-started framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkFramework.Url != url {
			t.Errorf("expected URL %s, got %s", url, zkFramework.Url)
		}
		if zkFramework.State != framework.Disconnected {
			t.Error("expected client to be disconnected")
		}
	})

	t.Run("Stop a non-started framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkFramework.Url != url {
			t.Errorf("expected URL %s, got %s", url, zkFramework.Url)
		}
		if zkFramework.State != framework.Disconnected {
			t.Error("expected client to be disconnected")
		}

		if err := zkFramework.Stop(); err != nil {
			if !framework.IsFrameworkNotYetStarted(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Wait a non-started framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkFramework.Url != url {
			t.Errorf("expected URL %s, got %s", url, zkFramework.Url)
		}
		if zkFramework.State != framework.Disconnected {
			t.Error("expected client to be disconnected")
		}

		if err := zkFramework.WaitConnection(5 * time.Second); err != nil {
			if !framework.IsFrameworkNotYetStarted(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Create and start the ZK framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
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
		if zkFramework.State != framework.Connected {
			t.Error("expected client to be connected")
		}
	})

	t.Run("Create and start twice the ZK framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Start(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Start(); err != nil {
			if !framework.IsFrameworkAlreadyStarted(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
		defer zkFramework.Stop()

		err = zkFramework.WaitConnection(10 * time.Second)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkFramework.State != framework.Connected {
			t.Error("expected client to be connected")
		}
	})

	t.Run("Create and start the ZK framework with connection timeout", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Start(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		defer zkFramework.Stop()

		err = zkFramework.WaitConnection(0 * time.Second)
		if err != nil && !framework.IsConnectionTimeout(err) {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})
}
