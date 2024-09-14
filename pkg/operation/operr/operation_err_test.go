package operr_test

import (
	"errors"
	"testing"

	"github.com/morphy76/zk/pkg/operation/operr"
)

func TestIsFrameworkNotReady(t *testing.T) {
	err := operr.ErrFrameworkNotReady
	if !operr.IsFrameworkNotReady(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsFrameworkNotReadyFalse(t *testing.T) {
	err := errors.New("some error")
	if operr.IsFrameworkNotReady(err) {
		t.Errorf("expected false, got true")
	}
}
