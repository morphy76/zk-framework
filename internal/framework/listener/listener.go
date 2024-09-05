/*
Package listener provides interfaces for listening to Zookeeper connection status changes and shutdown events.
*/
package listener

import (
	"errors"

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
StatusChangeListener is an interface for listening to Zookeeper connection status changes.
*/
type StatusChangeListener interface {
	UUID() string
	OnStatusChange(previous zk.State, current zk.State) error
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
ShutdownListener is an interface for listening to Zookeeper client shutdown events.
*/
type ShutdownListener interface {
	UUID() string
	OnShutdown() error
}

/*
ShutdownHandler is an interface for shutting down the Zookeeper client.
*/
type ShutdownHandler interface {
	AddShutdownListener(listener ShutdownListener) error
	RemoveShutdownListener(listener ShutdownListener) error
	NotifyShutdown()
}
