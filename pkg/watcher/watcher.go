/*
Package watcher provides a way to watch for changes in Zookeeper nodes.
*/
package watcher

import (
	"errors"
	"fmt"
	"path"
	"slices"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/framework"
)

/*
ErrUnknownNode is returned when the node is unknown.
*/
var ErrUnknownNode = errors.New("unknown node")

/*
IsUnknownNode checks if the error is ErrUnknownNode.
*/
func IsUnknownNode(err error) bool {
	return err == ErrUnknownNode
}

type watchListener struct {
	ID         string
	shutdownCh chan bool
}

func (w watchListener) UUID() string {
	return w.ID
}

func (w watchListener) OnShutdown() error {
	w.shutdownCh <- true
	return nil
}

/*
Set a watcher
*/
func Set(zkFramework framework.ZKFramework, nodeName string, outChan chan zk.Event, types ...zk.EventType) error {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, nodeName)...)
	if len(types) == 0 {
		types = []zk.EventType{
			zk.EventNodeDataChanged,
			zk.EventNodeChildrenChanged,
			zk.EventNodeCreated,
			zk.EventNodeDeleted,
		}
	}
	fmt.Printf("Set watcher at path %s for types %v\n", actualPath, types)

	cn := zkFramework.Cn()
	exists, _, out, err := cn.ExistsW(actualPath)
	if !exists {
		return ErrUnknownNode
	}
	if err != nil {
		return err
	}

	shutdown := make(chan bool)
	listener := watchListener{
		ID:         nodeName,
		shutdownCh: shutdown,
	}
	if err := zkFramework.AddShutdownListener(listener); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-shutdown:
				zkFramework.RemoveShutdownListener(listener)
				return
			case e := <-out:
				if slices.Contains(types, e.Type) {
					outChan <- e
				}
			}
		}
	}()

	return nil
}
