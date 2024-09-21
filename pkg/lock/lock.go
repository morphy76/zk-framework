/*
Package lock provides a simple lock mechanism for the application.
*/
package lock

import (
	"time"

	"github.com/morphy76/zk/pkg/core"
)

/*
Type is an enum that represents the type of lock.
*/
type Type int

const (
	// Unlocked is an unlocked lock.
	Unlocked Type = iota
	// RLock is a read lock.
	RLock
	// WLock is a write lock.
	WLock
)

/*
Lock is an interface that provides a simple lock mechanism for the application.
*/
type Lock interface {
	// RAcquire acquires a read lock on the lockable object.
	RAcquire(zkFramework core.ZKFramework, lockable Lockable, duration time.Duration) (func(), error)
	// WAcquire acquires a write lock on the lockable object.
	WAcquire(zkFramework core.ZKFramework, lockable Lockable, duration time.Duration) (func(), error)
	// HasLock checks if the lockable object has a lock.
	HasLock(zkFramework core.ZKFramework, lockable Lockable) (Type, error)
}

type lockImpl struct {
	lockspace string
}

/*
NewLock creates a new instance of the Lock interface.
*/
func NewLock(lockspace string) Lock {
	return &lockImpl{
		lockspace: lockspace,
	}
}

/*
RAcquire acquires a read lock on the lockable object.
*/
func (l *lockImpl) RAcquire(zkFramework core.ZKFramework, lockable Lockable, duration time.Duration) (func(), error) {
	return nil, nil
}

/*
WAcquire acquires a write lock on the lockable object.
*/
func (l *lockImpl) WAcquire(zkFramework core.ZKFramework, lockable Lockable, duration time.Duration) (func(), error) {
	return nil, nil
}

/*
HasLock checks if the lockable object has a lock.
*/
func (l *lockImpl) HasLock(zkFramework core.ZKFramework, lockable Lockable) (Type, error) {
	return Unlocked, nil
}
