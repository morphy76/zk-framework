package client_test

import (
	"os"
	"testing"

	"github.com/morphy76/zk/internal/client"
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

func TestZkClient(t *testing.T) {

	t.Run("Create a ZK framework with empty URL", func(t *testing.T) {
		_, err := client.CreateFramework("")
		if !client.IsInvalidConnectionURL(err) {
			t.Errorf("expected error %v, got %v", client.ErrInvalidConnectionURL, err)
		}
	})

	t.Run("Create a non-started framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkClient, err := client.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkClient.Url != url {
			t.Errorf("expected URL %s, got %s", url, zkClient.Url)
		}
		if zkClient.Connected {
			t.Error("expected client to be disconnected")
		}
	})

	t.Run("Create and start the ZK framework with valid URL", func(t *testing.T) {
		url := os.Getenv(zkHostEnv)
		zkClient, err := client.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		defer zkClient.Stop()

		if err := zkClient.Start(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if !zkClient.Connected {
			t.Error("expected client to be connected")
		}
	})
}
