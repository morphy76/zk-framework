/*
Package operr provides operation errors.
*/
package operr

import "errors"

/*
ErrFrameworkNotReady is returned when the framework is not ready.
*/
var ErrFrameworkNotReady = errors.New("framework not ready")

/*
IsFrameworkNotReady checks if the error is ErrFrameworkNotReady.
*/
func IsFrameworkNotReady(err error) bool {
	return err == ErrFrameworkNotReady
}
