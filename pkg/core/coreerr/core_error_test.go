package coreerr_test

import (
	"errors"
	"testing"

	"github.com/morphy76/zk/pkg/core/coreerr"
)

func TestIsListenerAlreadyExists(t *testing.T) {
	err := coreerr.ErrListenerAlreadyExists
	if !coreerr.IsListenerAlreadyExists(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsListenerNotFound(t *testing.T) {
	err := coreerr.ErrListenerNotFound
	if !coreerr.IsListenerNotFound(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsUnknownNode(t *testing.T) {
	err := coreerr.ErrUnknownNode
	if !coreerr.IsUnknownNode(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsListenerAlreadyExistsFalse(t *testing.T) {
	err := errors.New("some error")
	if coreerr.IsListenerAlreadyExists(err) {
		t.Errorf("expected false, got true")
	}
}

func TestIsListenerNotFoundFalse(t *testing.T) {
	err := errors.New("some error")
	if coreerr.IsListenerNotFound(err) {
		t.Errorf("expected false, got true")
	}
}

func TestIsUnknownNodeFalse(t *testing.T) {
	err := errors.New("some error")
	if coreerr.IsUnknownNode(err) {
		t.Errorf("expected false, got true")
	}
}
