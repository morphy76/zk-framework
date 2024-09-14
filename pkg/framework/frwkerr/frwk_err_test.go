package frwkerr_test

import (
	"errors"
	"testing"

	"github.com/morphy76/zk/pkg/framework/frwkerr"
)

func TestIsInvalidConnectionURL(t *testing.T) {
	err := frwkerr.ErrInvalidConnectionURL
	if !frwkerr.IsInvalidConnectionURL(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsConnectionTimeout(t *testing.T) {
	err := frwkerr.ErrConnectionTimeout
	if !frwkerr.IsConnectionTimeout(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsFrameworkAlreadyStarted(t *testing.T) {
	err := frwkerr.ErrFrameworkAlreadyStarted
	if !frwkerr.IsFrameworkAlreadyStarted(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsFrameworkNotYetStarted(t *testing.T) {
	err := frwkerr.ErrFrameworkNotYetStarted
	if !frwkerr.IsFrameworkNotYetStarted(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsInvalidConnectionURLFalse(t *testing.T) {
	err := errors.New("some error")
	if frwkerr.IsInvalidConnectionURL(err) {
		t.Errorf("expected false, got true")
	}
}

func TestIsConnectionTimeoutFalse(t *testing.T) {
	err := errors.New("some error")
	if frwkerr.IsConnectionTimeout(err) {
		t.Errorf("expected false, got true")
	}
}

func TestIsFrameworkAlreadyStartedFalse(t *testing.T) {
	err := errors.New("some error")
	if frwkerr.IsFrameworkAlreadyStarted(err) {
		t.Errorf("expected false, got true")
	}
}

func TestIsFrameworkNotYetStartedFalse(t *testing.T) {
	err := errors.New("some error")
	if frwkerr.IsFrameworkNotYetStarted(err) {
		t.Errorf("expected false, got true")
	}
}
