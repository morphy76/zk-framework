package client

import (
	"errors"
	"time"

	"github.com/go-zookeeper/zk"
)

/*
ErrInvalidConnectionURL is returned when the connection URL is invalid. A connection url is invalid when it is empty.
*/
var ErrInvalidConnectionURL = errors.New("invalid connection URL")

/*
IsInvalidConnectionURL checks if the error is an invalid connection URL error.
*/
func IsInvalidConnectionURL(err error) bool {
	return err == ErrInvalidConnectionURL
}

/*
ZKFramework represents a Zookeeper client with higher level capabilities, wrapping github.com/go-zookeeper/zk.
*/
type ZKFramework struct {
	Url       string
	Connected bool

	cn *zk.Conn
}

func (c *ZKFramework) Start() error {
	cn, _, err := zk.Connect([]string{c.Url}, 10*time.Second)
	c.cn = cn

	if err != nil {
		return err
	}
	c.Connected = true
	return nil
}

func (c *ZKFramework) Stop() {
	c.cn.Close()
	c.Connected = false
}

func CreateFramework(url string) (*ZKFramework, error) {
	if url == "" {
		return nil, ErrInvalidConnectionURL
	}

	return &ZKFramework{
		Url:       url,
		Connected: false,
	}, nil
}
