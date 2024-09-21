package lockerr_test

import (
	"errors"
	"testing"

	"github.com/morphy76/zk/pkg/lock/lockerr"
)

func TestIsErrSubjectEmpty(t *testing.T) {
	err := lockerr.ErrSubjectEmpty
	if !lockerr.IsSubjectEmpty(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsErrSubjectEmptyFalse(t *testing.T) {
	err := errors.New("some error")
	if lockerr.IsSubjectEmpty(err) {
		t.Errorf("expected false, got true")
	}
}
