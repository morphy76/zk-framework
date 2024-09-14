/*
Package frwkerr provides error types for the framework package.
*/
package frwkerr

import "errors"

/*
ErrInvalidConnectionURL is returned when the connection URL is invalid. A connection url is invalid when it is empty.
*/
var ErrInvalidConnectionURL = errors.New("invalid connection URL")

/*
ErrConnectionTimeout is returned when the connection to the Zookeeper server times out.
*/
var ErrConnectionTimeout = errors.New("connection timeout")

/*
ErrFrameworkAlreadyStarted is returned when the Zookeeper client is already started.
*/
var ErrFrameworkAlreadyStarted = errors.New("framework already started")

/*
ErrFrameworkNotYetStarted is returned when the Zookeeper client is not yet started.
*/
var ErrFrameworkNotYetStarted = errors.New("framework not yet started")

/*
IsInvalidConnectionURL checks if the error is an invalid connection URL error.
*/
func IsInvalidConnectionURL(err error) bool {
	return err == ErrInvalidConnectionURL
}

/*
IsConnectionTimeout checks if the error is a connection timeout error.
*/
func IsConnectionTimeout(err error) bool {
	return err == ErrConnectionTimeout
}

/*
IsFrameworkAlreadyStarted checks if the error is an already started error.
*/
func IsFrameworkAlreadyStarted(err error) bool {
	return err == ErrFrameworkAlreadyStarted
}

/*
IsFrameworkNotYetStarted checks if the error is a not yet started error.
*/
func IsFrameworkNotYetStarted(err error) bool {
	return err == ErrFrameworkNotYetStarted
}
