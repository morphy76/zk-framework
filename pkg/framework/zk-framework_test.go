package framework_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/morphy76/zk/internal/framework/listener"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/internal/test_util/mocks"
	"github.com/morphy76/zk/pkg/framework"
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

func TestZKFramework(t *testing.T) {

	t.Run("Create a ZK framework with empty URL", func(t *testing.T) {
		t.Log("Create a ZK framework with empty URL")
		_, err := framework.CreateFramework("")
		if !framework.IsInvalidConnectionURL(err) {
			t.Errorf("expected error %v, got %v", framework.ErrInvalidConnectionURL, err)
		}
	})

	t.Run("Create a non-started framework with valid URL", func(t *testing.T) {
		t.Log("Create a non-started framework with valid URL")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if zkFramework.URL() != url {
			t.Errorf("expected URL %s, got %s", url, zkFramework.URL())
		}
		if zkFramework.Connected() {
			t.Error("expected client to be disconnected")
		}
	})

	t.Run("Stop a non-started framework with valid URL", func(t *testing.T) {
		t.Log("Stop a non-started framework with valid URL")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Stop(); err != nil {
			if !framework.IsFrameworkNotYetStarted(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Wait a non-started framework with valid URL", func(t *testing.T) {
		t.Log("Wait a non-started framework with valid URL")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.WaitConnection(5 * time.Second); err != nil {
			if !framework.IsFrameworkNotYetStarted(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Create and start the ZK framework with valid URL", func(t *testing.T) {
		t.Log("Create and start the ZK framework with valid URL")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Start(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		defer func() {
			log.Println("Stopping ZK framework")
			zkFramework.Stop()
		}()

		err = zkFramework.WaitConnection(10 * time.Second)
		log.Println("Connection established")
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if !zkFramework.Connected() {
			t.Error(expectedClientToBeConnected)
		}
	})

	t.Run("Create and start twice the ZK framework with valid URL", func(t *testing.T) {
		t.Log("Create and start twice the ZK framework with valid URL")
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
		if !zkFramework.Connected() {
			t.Error(expectedClientToBeConnected)
		}
	})

	t.Run("Create and start the ZK framework with connection timeout", func(t *testing.T) {
		t.Log("Create and start the ZK framework with connection timeout")
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
		if err != nil {
			if !framework.IsConnectionTimeout(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Create and start the ZK framework with valid URL, waiting twice", func(t *testing.T) {
		t.Log("Create and start the ZK framework with valid URL, waiting twice")
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
		if !zkFramework.Connected() {
			t.Error(expectedClientToBeConnected)
		}

		err = zkFramework.WaitConnection(10 * time.Second)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
		if !zkFramework.Connected() {
			t.Error(expectedClientToBeConnected)
		}
	})

	t.Run("Create and start-stop the framework twice", func(t *testing.T) {
		t.Log("Create and start/stop the framework twice")
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

		if !zkFramework.Connected() {
			t.Error(expectedClientToBeConnected)
		}

		if err := zkFramework.Stop(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.Start(); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		err = zkFramework.WaitConnection(10 * time.Second)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if !zkFramework.Connected() {
			t.Error(expectedClientToBeConnected)
		}
	})

	t.Run("Test empty namespace", func(t *testing.T) {
		t.Log("Test empty namespace")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if zkFramework.Namespace() != "/" {
			t.Errorf("expected root namespace, got %s", zkFramework.Namespace())
		}
	})

	t.Run("Test non-empty namespace", func(t *testing.T) {
		t.Log("Test non-empty namespace")
		ns := uuid.New().String()
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url, ns)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if zkFramework.Namespace() != "/"+ns {
			t.Errorf("expected /%s namespace, got %s", ns, zkFramework.Namespace())
		}
	})

	t.Run("Test multiple namespaces", func(t *testing.T) {
		t.Log("Test multiple namespaces")
		ns1 := uuid.New().String()
		ns2 := uuid.New().String()
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url, ns1, ns2)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if zkFramework.Namespace() != "/"+ns1+"/"+ns2 {
			t.Errorf("expected /%s/%s namespace, got %s", ns1, ns2, zkFramework.Namespace())
		}
	})

	t.Run("Test multiple namespaces with trailing slash", func(t *testing.T) {
		t.Log("Test multiple namespaces with trailing slash")
		ns1 := uuid.New().String()
		ns2 := uuid.New().String()
		ns3 := uuid.New().String()
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url, "/"+ns1, "/"+ns2+"/", ns3+"/")
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if zkFramework.Namespace() != "/"+ns1+"/"+ns2+"/"+ns3 {
			t.Errorf("expected /%s/%s/%s namespace, got %s", ns1, ns2, ns3, zkFramework.Namespace())
		}
	})

	t.Run("Add a new status change listener", func(t *testing.T) {
		t.Log("Add a new status change listener")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedStatusChangeListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddStatusChangeListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("Add the same status change listener twice", func(t *testing.T) {
		t.Log("Add the same status change listener twice")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedStatusChangeListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddStatusChangeListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.AddStatusChangeListener(mockedListener); err != nil {
			if !listener.IsListenerAlreadyExists(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Remove a status change listener", func(t *testing.T) {
		t.Log("Remove a status change listener")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedStatusChangeListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddStatusChangeListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.RemoveStatusChangeListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("Remove a non-existent status change listener", func(t *testing.T) {
		t.Log("Remove a non-existent status change listener")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedStatusChangeListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.RemoveStatusChangeListener(mockedListener); err != nil {
			if !listener.IsListenerNotFound(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Test interaction when notifying status change listeners", func(t *testing.T) {
		t.Log("Test interaction when notifying status change listeners")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedStatusChangeListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddStatusChangeListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		zkFramework.NotifyStatusChange()
		if mockedListener.Interactions != 1 {
			t.Errorf("expected 1 interaction, got %d", mockedListener.Interactions)
		}
	})

	t.Run("Add a new shutdown listener", func(t *testing.T) {
		t.Log("Add a new shutdown listener")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedShutdownListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddShutdownListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("Add the same shutdown listener twice", func(t *testing.T) {
		t.Log("Add the same shutdown listener twice")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedShutdownListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddShutdownListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.AddShutdownListener(mockedListener); err != nil {
			if !listener.IsListenerAlreadyExists(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})

	t.Run("Remove a shutdown listener", func(t *testing.T) {
		t.Log("Remove a shutdown listener")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedShutdownListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.AddShutdownListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		if err := zkFramework.RemoveShutdownListener(mockedListener); err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}
	})

	t.Run("Remove a non-existent shutdown listener", func(t *testing.T) {
		t.Log("Remove a non-existent shutdown listener")
		url := os.Getenv(zkHostEnv)
		zkFramework, err := framework.CreateFramework(url)
		if err != nil {
			t.Errorf(unexpectedErrorFmt, err)
		}

		mockedListener := &mocks.MockedShutdownListener{
			ID:           uuid.New().String(),
			Interactions: 0,
		}
		if err := zkFramework.RemoveShutdownListener(mockedListener); err != nil {
			if !listener.IsListenerNotFound(err) {
				t.Errorf(unexpectedErrorFmt, err)
			}
		}
	})
}
