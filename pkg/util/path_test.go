package util_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/morphy76/zk/pkg/util"
)

const (
	expectedMessageFtm = "expected %s, got %s"
)

func TestZKOperation(t *testing.T) {

	t.Run("Concat simple paths", func(t *testing.T) {
		t.Log("Concat paths")
		p1 := uuid.New().String()
		p2 := uuid.New().String()
		p3 := uuid.New().String()
		expected := "/" + p1 + "/" + p2 + "/" + p3
		actual := util.ConcatPaths(p1, p2, p3)
		if actual != expected {
			t.Errorf(expectedMessageFtm, expected, actual)
		}
	})

	t.Run("Concat paths having slashes", func(t *testing.T) {
		t.Log("Concat paths having slashes")
		p1 := uuid.New().String()
		p2 := uuid.New().String()
		p3 := uuid.New().String()
		expected := "/" + p1 + "/" + p2 + "/" + p3
		actual := util.ConcatPaths("/"+p1, "/"+p2+"/", p3+"/")
		if actual != expected {
			t.Errorf(expectedMessageFtm, expected, actual)
		}
	})

	t.Run("Concat empty paths", func(t *testing.T) {
		t.Log("Concat empty paths")
		expected := "/"
		actual := util.ConcatPaths()
		if actual != expected {
			t.Errorf(expectedMessageFtm, expected, actual)
		}
	})

	t.Run("Concat multiple empty paths", func(t *testing.T) {
		t.Log("Concat multiple empty paths")
		expected := "/"
		actual := util.ConcatPaths("", "", "")
		if actual != expected {
			t.Errorf(expectedMessageFtm, expected, actual)
		}
	})
}
