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
ErrInvalidEvictionPolicy is returned when an invalid eviction policy is provided.
*/
var ErrInvalidEvictionPolicy = errors.New("invalid eviction policy")

/*
IsInvalidCacheSize returns true if the error is an ErrInvalidCacheSize.
*/
func IsInvalidCacheSize(err error) bool {
	return err == ErrInvalidCacheSize
}

/*
IsInvalidEvictionPolicy returns true if the error is an ErrInvalidEvictionPolicy.
*/
func IsInvalidEvictionPolicy(err error) bool {
	return err == ErrInvalidEvictionPolicy
}
