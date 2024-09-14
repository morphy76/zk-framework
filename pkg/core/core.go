/*
Package core provides the core interfaces for the Zookeeper framework.
*/
package core

import (
	"time"

	"github.com/go-zookeeper/zk"
)

/*
ZKFramework represents a Zookeeper client with higher level capabilities, wrapping github.com/go-zookeeper/zk.
*/
type ZKFramework interface {
	StatusChangeHandler
	ShutdownHandler
	Namespace() string
	Cn() *zk.Conn
	URL() string
	Started() bool
	Connected() bool
	Start() error
	WaitConnection(timeout time.Duration) error
	Stop() error
}

/*
StatusChangeHandler is an interface for listening to Zookeeper connection status changes.
*/
type StatusChangeHandler interface {
	AddStatusChangeListener(listener StatusChangeListener) error
	RemoveStatusChangeListener(listener StatusChangeListener) error
	NotifyStatusChange()
}

/*
ShutdownHandler is an interface for shutting down the Zookeeper client.
*/
type ShutdownHandler interface {
	AddShutdownListener(listener ShutdownListener) error
	RemoveShutdownListener(listener ShutdownListener) error
	NotifyShutdown()
}

/*
StatusChangeListener is an interface for listening to Zookeeper connection status changes.
*/
type StatusChangeListener interface {
	UUID() string
	OnStatusChange(zkFramework ZKFramework, previous zk.State, current zk.State) error
	Stop()
}

/*
ShutdownListener is an interface for listening to Zookeeper client shutdown events.
*/
type ShutdownListener interface {
	UUID() string
	OnShutdown(zkFramework ZKFramework) error
	Stop()
}
