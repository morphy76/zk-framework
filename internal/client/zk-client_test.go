package client_test

import (
	"context"
	"testing"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestZkClient(t *testing.T) {
	req := testcontainers.ContainerRequest{
		Image:        "zookeeper:3.9",
		ExposedPorts: []string{"2181/tcp"},
		WaitingFor:   testcontainers.WaitForLog("binding to port 2181"),
	}
	zkC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer zkC.Terminate(context.Background())
	// zkC.Host() returns the host where the container is running
	// zkC.MappedPort("2181") returns the mapped port
	// zkC.GetPort("2181/tcp") returns the original port
}
