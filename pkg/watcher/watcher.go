/*
Package watcher provides a way to watch for changes in Zookeeper nodes.
*/
package watcher

import (
	"fmt"
	"log"
	"path"
	"slices"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/core/coreerr"
)

type watchListener struct {
	ID         string
	shutdownCh chan bool
}

func (w watchListener) UUID() string {
	return w.ID
}

func (w watchListener) OnShutdown(zkFramework core.ZKFramework) error {
	w.shutdownCh <- true
	return nil
}

/*
Set a watcher
*/
func Set(zkFramework core.ZKFramework, nodeName string, outChan chan zk.Event, types ...zk.EventType) error {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, nodeName)...)
	if len(types) == 0 {
		types = []zk.EventType{
			zk.EventNodeDataChanged,
			zk.EventNodeChildrenChanged,
			zk.EventNodeCreated,
			zk.EventNodeDeleted,
		}
	}
	slices.Sort(types)

	nameParts := make([]string, 0, len(types)+1)
	for _, t := range types {
		nameParts = append(nameParts, fmt.Sprintf("%d", t))
	}
	nameParts = append(nameParts, nodeName)

	shutdown := make(chan bool)
	listener := watchListener{
		ID:         strings.Join(nameParts, "-"),
		shutdownCh: shutdown,
	}

	log.Printf("Set watcher at path %s for types %v with name %s\n", actualPath, types, listener.UUID())

	cn := zkFramework.Cn()
	exists, _, out, err := cn.ExistsW(actualPath)
	if !exists {
		return coreerr.ErrUnknownNode
	}
	if err != nil {
		return err
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
