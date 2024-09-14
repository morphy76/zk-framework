/*
Package operation provides operations on Zookeeper nodes.
*/
package operation

import (
	"context"
	"log"
	"path"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/core/coreerr"
	"github.com/morphy76/zk/pkg/framework/frwkerr"
)

type connectionConsumer[T any] func(*zk.Conn, chan T) error

/*
Ls lists the nodes at the given path.
*/
func Ls(zkFramework core.ZKFramework, paths ...string) ([]string, error) {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, paths...)...)
	log.Println("Listing nodes at path:", actualPath)

	outChan, errChan := execute(zkFramework, listNodes(actualPath))

	select {
	case out := <-outChan:
		return out, nil
	case err := <-errChan:
		return nil, err
	}
}

/*
CreateWithOptions creates a node at the given path with the given options.
*/
func CreateWithOptions(zkFramework core.ZKFramework, nodeName string, options CreateOptions) error {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, strings.Split(nodeName, "/")...)...)
	log.Println("Creating node at path:", actualPath)

	outChan, errChan := execute(zkFramework, createNode(actualPath, &options))

	select {
	case <-outChan:
		return nil
	case err := <-errChan:
		return err
	}
}

/*
Create creates a node at the given path.
*/
func Create(zkFramework core.ZKFramework, nodeName string) error {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, strings.Split(nodeName, "/")...)...)
	log.Println("Creating node at path:", actualPath)

	outChan, errChan := execute(zkFramework, createNode(actualPath, nil))

	path.Join()
	select {
	case <-outChan:
		return nil
	case err := <-errChan:
		return err
	}
}

/*
Exists checks if a node exists at the given path.
*/
func Exists(zkFramework core.ZKFramework, nodeName string) (bool, error) {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, strings.Split(nodeName, "/")...)...)
	log.Println("Checking if node exists at path:", actualPath)

	outChan, errChan := execute(zkFramework, existsNode(actualPath))

	select {
	case out := <-outChan:
		return out, nil
	case err := <-errChan:
		return false, err
	}
}

/*
Delete deletes a node at the given path.
*/
func Delete(zkFramework core.ZKFramework, nodeName string) error {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, strings.Split(nodeName, "/")...)...)
	log.Println("Deleting node at path:", actualPath)

	outChan, errChan := execute(zkFramework, deleteNode(actualPath))

	select {
	case <-outChan:
		return nil
	case err := <-errChan:
		return err
	}
}

/*
Update updates a node at the given path.
*/
func Update(zkFramework core.ZKFramework, nodeName string, data []byte) (int32, error) {
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, strings.Split(nodeName, "/")...)...)
	log.Println("Updating node at path:", actualPath)

	outChan, errChan := execute(zkFramework, updateNode(actualPath, data))

	select {
	case out := <-outChan:
		return out, nil
	case err := <-errChan:
		return 0, err
	}
}

/*
Get gets a node at the given path.
*/
func Get(zkFramework core.ZKFramework, nodeName string) ([]byte, error) {
	// TODO with stats
	actualPath := path.Join(append([]string{zkFramework.Namespace()}, strings.Split(nodeName, "/")...)...)
	log.Println("Getting node at path:", actualPath)

	outChan, errChan := execute(zkFramework, getNode(actualPath))

	select {
	case out := <-outChan:
		return out, nil
	case err := <-errChan:
		return nil, err
	}
}

func listNodes(path string) connectionConsumer[[]string] {
	return func(cn *zk.Conn, outChan chan []string) error {
		children, _, err := cn.Children(path)
		if err != nil {
			return err
		}
		outChan <- children
		return nil
	}
}

func createNode(path string, options *CreateOptions) connectionConsumer[bool] {
	return func(cn *zk.Conn, outChan chan bool) error {
		recursivelyGrantParent(path, cn)
		data, flag, acl := parseOptions(options)
		_, err := cn.Create(path, data, flag, acl)
		if err != nil {
			return err
		}
		outChan <- true
		return nil
	}
}

func parseOptions(options *CreateOptions) ([]byte, int32, []zk.ACL) {
	if options == nil {
		return []byte{}, 0, zk.WorldACL(zk.PermAll)
	}

	data := options.Data
	flag := options.Mode
	acl := options.ACL

	if data == nil {
		data = []byte{}
	}

	if flag == 0 {
		flag = 0
	}

	if acl == nil {
		acl = zk.WorldACL(zk.PermAll)
	}

	return data, flag, acl
}

func deleteNode(path string) connectionConsumer[bool] {
	return func(cn *zk.Conn, outChan chan bool) error {
		exists, _, err := cn.Exists(path)
		if err != nil {
			return err
		}

		if !exists {
			return coreerr.ErrUnknownNode
		}

		err = cn.Delete(path, -1)
		if err != nil {
			return err
		}
		outChan <- true
		return nil
	}
}

func updateNode(path string, data []byte) connectionConsumer[int32] {
	return func(cn *zk.Conn, outChan chan int32) error {
		exists, _, err := cn.Exists(path)
		if err != nil {
			return err
		}

		if !exists {
			return coreerr.ErrUnknownNode
		}

		stat, err := cn.Set(path, data, -1)
		if err != nil {
			return err
		}
		outChan <- stat.Version
		return nil
	}
}

func getNode(path string) connectionConsumer[[]byte] {
	return func(cn *zk.Conn, outChan chan []byte) error {
		data, _, err := cn.Get(path)
		if err != nil {
			return err
		}
		outChan <- data
		return nil
	}
}

func recursivelyGrantParent(nodeName string, cn *zk.Conn) error {
	parent := path.Dir(nodeName)
	if parent == "/" {
		return nil
	}

	exists, _, err := cn.Exists(parent)
	if err != nil {
		return err
	}

	if !exists {
		err := recursivelyGrantParent(parent, cn)
		if err != nil {
			return err
		}
		_, err = cn.Create(parent, []byte{}, zk.FlagContainer, zk.WorldACL(zk.PermAll))
		if err != nil {
			return err
		}
	}
	return nil
}

func existsNode(path string) connectionConsumer[bool] {
	return func(cn *zk.Conn, outChan chan bool) error {
		exists, _, err := cn.Exists(path)
		if err != nil {
			return err
		}
		outChan <- exists
		return nil
	}
}

func execute[T any](zkFramework core.ZKFramework, cnConsumer connectionConsumer[T]) (chan T, chan error) {

	outChan := make(chan T)
	errChan := make(chan error)

	if !zkFramework.Started() {
		errChan <- frwkerr.ErrFrameworkNotYetStarted
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		defer close(errChan)

		go func() {
			defer close(outChan)

			err := cnConsumer(zkFramework.Cn(), outChan)
			if err != nil {
				errChan <- err
			}
			cancel()
		}()

		<-ctx.Done()
		if ctx.Err() != nil {
			errChan <- ctx.Err()
		}

	}()

	return outChan, errChan
}
