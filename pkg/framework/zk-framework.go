package framework

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/util"
)

/*
ErrInvalidConnectionURL is returned when the connection URL is invalid. A connection url is invalid when it is empty.
*/
var ErrInvalidConnectionURL = errors.New("invalid connection URL")

/*
ErrConnectionTimeout is returned when the connection to the Zookeeper server times out.
*/
var ErrConnectionTimeout = errors.New("connection timeout")

/*
ErrFrameworkAlreadyStarted is returned when the Zookeeper client is already started.
*/
var ErrFrameworkAlreadyStarted = errors.New("framework already started")

/*
ErrFrameworkNotYetStarted is returned when the Zookeeper client is not yet started.
*/
var ErrFrameworkNotYetStarted = errors.New("framework not yet started")

/*
IsInvalidConnectionURL checks if the error is an invalid connection URL error.
*/
func IsInvalidConnectionURL(err error) bool {
	return err == ErrInvalidConnectionURL
}

/*
IsConnectionTimeout checks if the error is a connection timeout error.
*/
func IsConnectionTimeout(err error) bool {
	return err == ErrConnectionTimeout
}

/*
IsFrameworkAlreadyStarted checks if the error is an already started error.
*/
func IsFrameworkAlreadyStarted(err error) bool {
	return err == ErrFrameworkAlreadyStarted
}

/*
IsFrameworkNotYetStarted checks if the error is a not yet started error.
*/
func IsFrameworkNotYetStarted(err error) bool {
	return err == ErrFrameworkNotYetStarted
}

const (
	defaultReconnectionTimeoutMs = 100
)

/*
ZKFramework represents a Zookeeper client with higher level capabilities, wrapping github.com/go-zookeeper/zk.
*/
type ZKFramework struct {
	namespace     string
	url           string
	state         zk.State
	previousState zk.State
	started       bool

	cn                    *zk.Conn
	events                <-chan zk.Event
	reconnectionTimeoutMs uint64

	shutdown          chan bool
	shutdownConsumers int

	statusChange          chan zk.State
	statusChangeConsumers int
	statusChangeLock      sync.RWMutex
}

func (c *ZKFramework) Namespace() string {
	return c.namespace
}

func (c *ZKFramework) Cn() *zk.Conn {
	return c.cn
}

/*
Url returns the URL of the Zookeeper client.
*/
func (c *ZKFramework) Url() string {
	return c.url
}

/*
Started returns whether the Zookeeper client is started.
*/
func (c *ZKFramework) Started() bool {
	return c.started
}

/*
Connected returns whether the Zookeeper client is connected to the server.
*/
func (c *ZKFramework) Connected() bool {
	return isConnectedState(c.state)
}

/*
Start connects to the Zookeeper server and starts watching connection events.
*/
func (c *ZKFramework) Start() error {
	if c.started {
		return ErrFrameworkAlreadyStarted
	}

	log.Printf("connecting to Zookeeper server at %s", c.url)

	c.started = true

	return c.tryConnect()
}

/*
WaitConnection waits for the connection to the Zookeeper server to be established.
*/
func (c *ZKFramework) WaitConnection(timeout time.Duration) error {
	if !c.started {
		return ErrFrameworkNotYetStarted
	}

	c.statusChangeLock.Lock()
	defer c.statusChangeLock.Unlock()

	if c.Connected() {
		return nil
	}

	log.Printf("waiting for connection to Zookeeper server at %s", c.url)

	c.shutdownConsumers++
	defer func() {
		c.shutdownConsumers--
	}()

	c.statusChangeConsumers++
	defer func() {
		c.statusChangeConsumers--
	}()

	for {
		select {
		case state := <-c.statusChange:
			if state == zk.StateConnected {
				log.Printf("connected to Zookeeper server at %s", c.url)
				return nil
			}
		case <-c.shutdown:
			return nil
		case <-time.After(timeout):
			return ErrConnectionTimeout
		}
	}
}

/*
Stop closes the connection to the Zookeeper server.
*/
func (c *ZKFramework) Stop() error {
	if !c.started {
		return ErrFrameworkNotYetStarted
	}
	defer c.cn.Close()

	log.Printf("closing connection to Zookeeper server at %s", c.url)
	c.started = false

	c.stopBgTasks()

	c.state = zk.StateDisconnected

	return nil
}

func (c *ZKFramework) watchEvents() {
	log.Printf("watching events from Zookeeper server at %s", c.url)

	c.shutdownConsumers++
	defer func() {
		c.shutdownConsumers--
	}()

	for {
		select {
		case <-c.shutdown:
			return
		case event := <-c.events:
			for i := 0; i < c.statusChangeConsumers; i++ {
				c.statusChange <- event.State
			}
		}
	}
}

func (c *ZKFramework) connectionWatcher() {
	log.Printf("watching connection to Zookeeper server at %s", c.url)

	c.shutdownConsumers++
	defer func() {
		c.shutdownConsumers--
	}()

	c.statusChangeConsumers++
	defer func() {
		c.statusChangeConsumers--
	}()

	for {
		select {
		case <-c.shutdown:
			return
		case state := <-c.statusChange:
			c.handleStatusChange(state)
		}
	}
}

func (c *ZKFramework) handleStatusChange(state zk.State) {
	c.statusChangeLock.Lock()
	defer c.statusChangeLock.Unlock()

	if state == c.state {
		return
	}

	c.previousState = c.state
	c.state = state
	log.Printf("status change from %s to %s", c.previousState, c.state)
	if !c.previouslyConnected() && c.Connected() {
		c.reconnectionTimeoutMs = defaultReconnectionTimeoutMs
	}
	if c.started && c.previouslyConnected() && !c.Connected() {
		log.Printf("connection to Zookeeper server at %s lost, trying to reconnect", c.url)
		c.invalidateCn()
	}
}

func (c *ZKFramework) tryConnect() error {
	cn, events, err := zk.Connect([]string{c.url}, 10*time.Second)
	if err != nil {
		return err
	}
	c.cn = cn
	c.events = events
	go c.watchEvents()
	go c.connectionWatcher()

	return nil
}

func (c *ZKFramework) invalidateCn() {
	c.stopBgTasks()
	<-time.After(time.Duration(c.reconnectionTimeoutMs) * time.Millisecond)
	c.reconnectionTimeoutMs *= 2

	if c.cn != nil {
		c.cn.Close()
	}
	c.tryConnect()
}

func (c *ZKFramework) previouslyConnected() bool {
	return isConnectedState(c.previousState)
}

func (c *ZKFramework) stopBgTasks() {
	for i := 0; i < c.shutdownConsumers; i++ {
		c.shutdown <- true
	}
}

func isConnectedState(state zk.State) bool {
	return state == zk.StateConnected ||
		state == zk.StateHasSession ||
		state == zk.StateConnectedReadOnly ||
		state == zk.StateSaslAuthenticated
}

func CreateFramework(url string, namespace ...string) (*ZKFramework, error) {
	if url == "" {
		return nil, ErrInvalidConnectionURL
	}

	useNamespace := util.ConcatPaths(namespace...)

	return &ZKFramework{
		namespace: useNamespace,
		url:       url,
		state:     zk.StateDisconnected,
		started:   false,

		shutdownConsumers:     0,
		statusChangeConsumers: 0,
		reconnectionTimeoutMs: defaultReconnectionTimeoutMs,

		shutdown:         make(chan bool),
		statusChange:     make(chan zk.State),
		statusChangeLock: sync.RWMutex{},
	}, nil
}
