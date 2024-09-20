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

var watchListeners = make(map[string]*watchListener)

type watchListener struct {
	ID           string
	path         string
	shutdownCh   chan bool
	outCh        chan zk.Event
	types        []zk.EventType
	watching     bool
	disconnected bool
}

func (w watchListener) UUID() string {
	return w.ID
}

func (w *watchListener) OnShutdown(zkFramework core.ZKFramework) error {
	log.Printf("Watcher %s: OnShutdown\n", w.ID)
	if !w.watching {
		return nil
	}
	w.Stop()
	return nil
}

func (w *watchListener) OnStatusChange(zkFramework core.ZKFramework, previous zk.State, current zk.State) error {
	log.Printf("Watcher %s: State change from %s to %s\n", w.ID, previous, current)
	if w.watching {
		if !w.disconnected && !zkFramework.Connected() {
			log.Printf("Watcher %s: Connection lost\n", w.ID)
			w.disconnected = true
			w.shutdownCh <- true
		}
		if w.disconnected && zkFramework.Connected() {
			log.Printf("Watcher %s: Connection established\n", w.ID)
			w.Start(zkFramework)
			w.disconnected = false
		}

	}
	return nil
}

func (w *watchListener) Start(zkFramework core.ZKFramework) error {
	log.Printf("Watcher %v: Start\n", w)

	cn := zkFramework.Cn()
	exists, _, out, err := cn.ExistsW(w.path)
	if !exists {
		return coreerr.ErrUnknownNode
	}
	if err != nil {
		return err
	}
	watchFn := func() {
		for {
			select {
			case <-w.shutdownCh:
				log.Printf("Watcher %s: Shutdown\n", w.ID)
				return
			case e := <-out:
				if slices.Contains(w.types, e.Type) {
					w.outCh <- e
				}
			}
		}
	}

	w.watching = true
	go watchFn()
	return nil
}

func (w *watchListener) Stop() {
	log.Printf("Watcher %v: Stop\n", w)
	w.watching = false
	w.shutdownCh <- true
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
	nameParts = append(nameParts, actualPath)

	id := namePartsToID(nameParts)
	watchListeners[id] = &watchListener{
		ID:         id,
		shutdownCh: make(chan bool),
		outCh:      outChan,
		path:       actualPath,
		types:      types,
	}
	log.Printf("Set watcher listener at path %s for types %v with name %s\n", actualPath, types, watchListeners[id].UUID())

	if err := zkFramework.AddShutdownListener(watchListeners[id]); err != nil {
		return err
	}
	if err := zkFramework.AddStatusChangeListener(watchListeners[id]); err != nil {
		zkFramework.RemoveShutdownListener(watchListeners[id])
		return err
	}

	err := watchListeners[id].Start(zkFramework)

	return err
}

/*
UnSet a watcher
*/
func UnSet(zkFramework core.ZKFramework, nodeName string, types ...zk.EventType) error {
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
	nameParts = append(nameParts, actualPath)

	id := namePartsToID(nameParts)

	watchListeners[id].Stop()
	if err := zkFramework.RemoveShutdownListener(watchListeners[id]); err != nil {
		log.Printf("Error removing shutdown listener: %s\n", err)
	}
	if err := zkFramework.RemoveStatusChangeListener(watchListeners[id]); err != nil {
		log.Printf("Error removing status change listener: %s\n", err)
	}
	delete(watchListeners, id)
	return nil
}

func namePartsToID(nameParts []string) string {
	return strings.Join(nameParts, "-")
}
