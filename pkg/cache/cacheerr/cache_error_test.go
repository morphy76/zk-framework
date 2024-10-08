package cacheerr_test

import (
	"errors"
	"testing"

	"github.com/morphy76/zk/pkg/cache/cacheerr"
)

func TestIsErrInvalidCacheSize(t *testing.T) {
	err := cacheerr.ErrInvalidCacheSize
	if !cacheerr.IsInvalidCacheSize(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsErrInvalidCacheSizeFalse(t *testing.T) {
	err := errors.New("some error")
	if cacheerr.IsInvalidCacheSize(err) {
		t.Errorf("expected false, got true")
	}
}

func TestIsErrInvalidEvictionPolicy(t *testing.T) {
	err := cacheerr.ErrInvalidEvictionPolicy
	if !cacheerr.IsInvalidEvictionPolicy(err) {
		t.Errorf("expected true, got false")
	}
}

func TestIsErrInvalidEvictionPolicyFalse(t *testing.T) {
	err := errors.New("some error")
	if cacheerr.IsInvalidEvictionPolicy(err) {
		t.Errorf("expected false, got true")
	}
}
