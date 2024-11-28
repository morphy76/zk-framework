package lock

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/morphy76/zk/pkg/lock/lockerr"
)

const noReason = ""

/*
LockableBuilder is a builder for the Lockable interface.
*/
type LockableBuilder struct {
	subject string
	reason  string
}

/*
Lockable is an interface that can be implemented by any type that can be locked.
*/
type Lockable interface {
	// Hash returns a unique hash for the lockable object.
	Hash() string
}

type lockableImpl struct {
	lockHash string
}

func (l *lockableImpl) Hash() string {
	return l.lockHash
}

/*
NewLockableBuilder creates a new instance of the LockableBuilder.
*/
func NewLockableBuilder() LockableBuilder {
	return LockableBuilder{
		subject: "",
		reason:  noReason,
	}
}

/*
WithSubject sets the subject for the lockable object.
*/
func (b LockableBuilder) WithSubject(subject string) LockableBuilder {
	b.subject = subject
	return b
}

/*
WithReason sets the reason for the lockable object.
*/
func (b LockableBuilder) WithReason(reason string) LockableBuilder {
	b.reason = reason
	return b
}

/*
Build creates a new instance of the Lockable interface.
*/
func (b LockableBuilder) Build() (Lockable, error) {

	if b.subject == "" {
		return nil, lockerr.ErrSubjectEmpty
	}

	useLockableID := b.subject + b.reason
	lockableHash := sha256.Sum256([]byte(useLockableID))
	encodedHash := base64.StdEncoding.EncodeToString(lockableHash[:])

	return &lockableImpl{
		lockHash: encodedHash,
	}, nil
}
