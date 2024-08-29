package operation

import (
	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/internal/framework"
)

/*
ZKCnConsumer is a function that consumes a Zookeeper connection.
*/
type ZKCnConsumer func(*zk.Conn) error

/*
ZKOperation is a function that performs an operation on a Zookeeper client.
*/
type ZKOperation func(*framework.ZKFramework) ZKCnConsumer

var Ls ZKOperation = func(zkFramework *framework.ZKFramework) ZKCnConsumer {
	return nil
}
