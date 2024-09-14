/*
Package testutil provides utilities for testing.
*/
package testutil

import (
	"context"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	image       = "zookeeper:3.9"
	exposedPort = "2181/tcp"
)

/*
StartTestServer starts a Zookeeper test server.

Returns:
  - testcontainers.Container: the Zookeeper container
  - context.Context: the context
  - error: the error
*/
func StartTestServer() (testcontainers.Container, context.Context, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{exposedPort},
		WaitingFor:   wait.ForListeningPort(exposedPort),
	}
	zkC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, nil, err
	}

	err = zkC.Start(ctx)

	return zkC, ctx, err
}
