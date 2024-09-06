/*
Package core provides the core interfaces for the Zookeeper framework.
*/
package core

import (
	"errors"
	"time"

	"github.com/go-zookeeper/zk"
)

/*
ErrListenerAlreadyExists is returned when a listener is already added to the listener list.
*/
var ErrListenerAlreadyExists = errors.New("listener already exists")

/*
ErrListenerNotFound is returned when a listener is not found in the listener list.
*/
var ErrListenerNotFound = errors.New("listener not found")

/*
IsListenerAlreadyExists checks if the error is a listener already exists error.
*/
func IsListenerAlreadyExists(err error) bool {
	return err == ErrListenerAlreadyExists
}

/*
IsListenerNotFound checks if the error is a listener not found error.
*/
func IsListenerNotFound(err error) bool {
	return err == ErrListenerNotFound
}

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
}

/*
ShutdownListener is an interface for listening to Zookeeper client shutdown events.
*/
type ShutdownListener interface {
	UUID() string
	OnShutdown(zkFramework ZKFramework) error
}
