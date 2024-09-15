package operation_test

import (
	"testing"

	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
	testutil "github.com/morphy76/zk/internal/test_util"
	"github.com/morphy76/zk/pkg/operation"
)

func TestDefaultCreateOptionsBuilder(t *testing.T) {
	opts := operation.NewCreateOptionsBuilder().Build()

	if opts.ACL != nil {
		t.Errorf("Expected ACL to be nil, got %v", opts.ACL)
	}

	if opts.Data != nil {
		t.Errorf("Expected Data to be nil, got %v", opts.Data)
	}

	if opts.Mode != 0 {
		t.Errorf("Expected Mode to be 0, got %v", opts.Mode)
	}
}

func TestCreateOptionsBuilder(t *testing.T) {
	acl := zk.WorldACL(zk.PermAll)
	data := []byte(uuid.New().String())
	mode := int32(zk.FlagEphemeral)

	opts := operation.NewCreateOptionsBuilder().
		WithACL(acl).
		WithData(data).
		WithMode(mode).
		Build()

	if opts.ACL == nil {
		t.Errorf("Expected ACL to be %v, got nil", acl)
	}
	if opts.ACL[0].Perms != acl[0].Perms {
		t.Errorf("Expected ACL to be %v, got %v", acl, opts.ACL)
	}
	if opts.Data == nil {
		t.Errorf("Expected Data to be %v, got nil", data)
	}
	if string(opts.Data) != string(data) {
		t.Errorf("Expected Data to be %v, got %v", data, opts.Data)
	}
	if opts.Mode != mode {
		t.Errorf("Expected Mode to be %v, got %v", mode, opts.Mode)
	}
}

func TestCreateNodeWithOptions(t *testing.T) {
	t.Log("Create node with options")
	zkFramework, err := testutil.ConnectFramework()
	if err != nil {
		t.Errorf(unexpectedErrorFmt, err)
	}
	defer zkFramework.Stop()

	nodeName := uuid.New().String()
	acl := zk.WorldACL(zk.PermAll)
	data := []byte(uuid.New().String())
	mode := int32(zk.FlagEphemeral)

	opts := operation.NewCreateOptionsBuilder().
		WithACL(acl).
		WithData(data).
		WithMode(mode).
		Build()

	if err := operation.CreateWithOptions(zkFramework, nodeName, opts); err != nil {
		t.Errorf(unexpectedErrorFmt, err)
	}

	readData, err := operation.Get(zkFramework, nodeName)
	if err != nil {
		t.Errorf(unexpectedErrorFmt, err)
	}
	if string(readData) != string(data) {
		t.Errorf("expected data to be %s, got %s", string(data), string(readData))
	}
}
