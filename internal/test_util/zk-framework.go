package testutil

import (
	"os"
	"time"

	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/framework"
)

const (
	zkHostEnv                   = "ZK_HOST"
	unexpectedErrorFmt          = "unexpected error %v"
	expectedClientToBeConnected = "expected client to be connected"
)

/*
ConnectFramework creates a new ZKFramework instance and connects to the Zookeeper server.
*/
func ConnectFramework() (core.ZKFramework, error) {
	url := os.Getenv(zkHostEnv)
	zkFramework, err := framework.CreateFramework(url)
	if err != nil {
		return nil, err
	}

	if err := zkFramework.Start(); err != nil {
		return nil, err
	}

	err = zkFramework.WaitConnection(10 * time.Second)
	if err != nil {
		return nil, err
	}
	if !zkFramework.Connected() {
		return nil, err
	}
	return zkFramework, nil
}
