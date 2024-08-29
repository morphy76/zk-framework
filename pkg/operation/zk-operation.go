package operation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/framework"
	"github.com/morphy76/zk/pkg/util"
)

/*
ErrFrameworkNotReady is returned when the framework is not ready.
*/
var ErrFrameworkNotReady = errors.New("framework not ready")

/*
IsFrameworkNotReady checks if the error is ErrFrameworkNotReady.
*/
func IsFrameworkNotReady(err error) bool {
	return err == ErrFrameworkNotReady
}

type connectionConsumer[T any] func(*zk.Conn, chan T) error

/*
Ls lists the nodes at the given path.
*/
func Ls(zkFramework *framework.ZKFramework, paths ...string) ([]string, error) {
	actualPath := util.ConcatPaths(append([]string{zkFramework.Namespace()}, paths...)...)
	fmt.Println("Listing nodes at path:", actualPath)

	outChan, errChan := execute(zkFramework, listNodes(actualPath))

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

func execute[T any](zkFramework *framework.ZKFramework, cnConsumer connectionConsumer[T]) (chan T, chan error) {

	outChan := make(chan T)
	errChan := make(chan error)

	if !zkFramework.Started() {
		errChan <- framework.ErrFrameworkNotYetStarted
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {

		go func() {
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
