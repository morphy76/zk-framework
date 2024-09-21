package lock_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/morphy76/zk/pkg/lock"
	"github.com/morphy76/zk/pkg/lock/lockerr"
)

func TestDefaultLockableBuilderWithoutSubject(t *testing.T) {
	_, err := lock.NewLockableBuilder().
		Build()
	if !lockerr.IsSubjectEmpty(err) {
		t.Errorf("Expected error to be ErrSubjectEmpty, got %v", err)
	}
}

func TestDefaultLockableBuilder(t *testing.T) {
	lockable, err := lock.NewLockableBuilder().
		WithSubject(uuid.New().String()).
		Build()
	if lockerr.IsSubjectEmpty(err) {
		t.Errorf("Expected error to be nil, got %v", err)
	}
	if lockable.Hash() == "" {
		t.Errorf("Expected Hash to be non-empty, got empty")
	}
}
