/*
Package coreerr provides error types for the core package.
*/
package coreerr

import "errors"

/*
ErrListenerAlreadyExists is returned when a listener is already added to the listener list.
*/
var ErrListenerAlreadyExists = errors.New("listener already exists")

/*
ErrListenerNotFound is returned when a listener is not found in the listener list.
*/
var ErrListenerNotFound = errors.New("listener not found")

/*
ErrUnknownNode is returned when the node is unknown.
*/
var ErrUnknownNode = errors.New("unknown node")

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
IsUnknownNode checks if the error is ErrUnknownNode.
*/
func IsUnknownNode(err error) bool {
	return err == ErrUnknownNode
}
