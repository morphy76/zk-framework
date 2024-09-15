/*
Package cacheerr provides error types for cache operations.
*/
package cacheerr

import "errors"

/*
ErrInvalidCacheSize is returned when an invalid cache size is provided.
*/
var ErrInvalidCacheSize = errors.New("invalid cache size")

/*
IsInvalidCacheSize returns true if the error is an ErrInvalidCacheSize.
*/
func IsInvalidCacheSize(err error) bool {
	return err == ErrInvalidCacheSize
}
