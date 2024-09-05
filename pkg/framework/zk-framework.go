/*
Package framework provides a higher level Zookeeper client with more capabilities than the standard Zookeeper client.
*/
package framework

import (
	"errors"
	"log"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/internal/framework/listener"
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
type ZKFramework interface {
	listener.StatusChangeHandler
	listener.ShutdownHandler
	Namespace() string
	Cn() *zk.Conn
	URL() string
	Started() bool
	Connected() bool
	Start() error
	WaitConnection(timeout time.Duration) error
	Stop() error
}

type zKFrameworkImpl struct {
	namespace     string
	url           string
	state         zk.State
	previousState zk.State
	started       bool

	cn                    *zk.Conn
	events                <-chan zk.Event
	reconnectionTimeoutMs uint64

	shutdown          chan bool
	shutdownConsumers atomic.Int32
	shutdownListeners map[string]listener.ShutdownListener

	statusChange          chan zk.State
	statusChangeConsumers atomic.Int32
	statusChangeLock      sync.RWMutex
	statusChangeListeners map[string]listener.StatusChangeListener
}

func (c *zKFrameworkImpl) Namespace() string {
	return c.namespace
}

func (c *zKFrameworkImpl) Cn() *zk.Conn {
	return c.cn
}

/*
Url returns the URL of the Zookeeper client.
*/
func (c *zKFrameworkImpl) URL() string {
	return c.url
}

/*
Started returns whether the Zookeeper client is started.
*/
func (c *zKFrameworkImpl) Started() bool {
	return c.started
}

/*
Connected returns whether the Zookeeper client is connected to the server.
*/
func (c *zKFrameworkImpl) Connected() bool {
	c.statusChangeLock.RLock()
	defer c.statusChangeLock.RUnlock()
	return isConnectedState(c.state)
}

/*
Start connects to the Zookeeper server and starts watching connection events.
*/
func (c *zKFrameworkImpl) Start() error {
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
func (c *zKFrameworkImpl) WaitConnection(timeout time.Duration) error {
	if !c.started {
		return ErrFrameworkNotYetStarted
	}

	if c.Connected() {
		return nil
	}

	log.Printf("waiting for connection to Zookeeper server at %s", c.url)

	c.shutdownConsumers.Add(1)
	defer func() {
		c.shutdownConsumers.Add(-1)
	}()

	c.statusChangeConsumers.Add(1)
	defer func() {
		c.statusChangeConsumers.Add(-1)
	}()

	for {
		select {
		case <-c.statusChange:
			if c.Connected() {
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
func (c *zKFrameworkImpl) Stop() error {
	c.statusChangeLock.Lock()
	defer c.statusChangeLock.Unlock()

	if !c.started {
		return ErrFrameworkNotYetStarted
	}
	defer c.cn.Close()

	log.Printf("closing connection to Zookeeper server at %s", c.url)
	c.started = false

	c.stopBgTasks()
	c.NotifyShutdown()

	c.state = zk.StateDisconnected

	return nil
}

/*
AddStatusChangeListener adds a listener for Zookeeper connection status changes.
*/
func (c *zKFrameworkImpl) AddStatusChangeListener(statusChangeListener listener.StatusChangeListener) error {
	// TODO locks

	if found := c.statusChangeListeners[statusChangeListener.UUID()]; found != nil {
		return listener.ErrListenerAlreadyExists
	}

	c.statusChangeListeners[statusChangeListener.UUID()] = statusChangeListener
	return nil
}

/*
RemoveStatusChangeListener removes a listener for Zookeeper connection status changes.
*/
func (c *zKFrameworkImpl) RemoveStatusChangeListener(statusChangeListener listener.StatusChangeListener) error {
	// TODO locks

	if found := c.statusChangeListeners[statusChangeListener.UUID()]; found == nil {
		return listener.ErrListenerNotFound
	}

	delete(c.statusChangeListeners, statusChangeListener.UUID())
	return nil
}

/*
NotifyStatusChange notifies all listeners of a Zookeeper connection status change.
*/
func (c *zKFrameworkImpl) NotifyStatusChange() {
	// TODO locks

	for _, listener := range c.statusChangeListeners {
		if err := listener.OnStatusChange(c.previousState, c.state); err != nil {
			log.Printf("error notifying status change listener: %s", err)
		}
	}
}

/*
AddShutdownListener adds a listener for Zookeeper client shutdown events.
*/
func (c *zKFrameworkImpl) AddShutdownListener(shutdownListener listener.ShutdownListener) error {
	if found := c.shutdownListeners[shutdownListener.UUID()]; found != nil {
		return listener.ErrListenerAlreadyExists
	}

	c.shutdownListeners[shutdownListener.UUID()] = shutdownListener
	return nil
}

/*
RemoveShutdownListener removes a listener for Zookeeper client shutdown events.
*/
func (c *zKFrameworkImpl) RemoveShutdownListener(shutdownListener listener.ShutdownListener) error {
	if found := c.shutdownListeners[shutdownListener.UUID()]; found == nil {
		return listener.ErrListenerNotFound
	}

	delete(c.shutdownListeners, shutdownListener.UUID())
	return nil
}

/*
NotifyShutdown notifies all listeners of a Zookeeper client shutdown event.
*/
func (c *zKFrameworkImpl) NotifyShutdown() {
	for _, listener := range c.shutdownListeners {
		if err := listener.OnShutdown(); err != nil {
			log.Printf("error notifying shutdown listener: %s", err)
		}
	}
}

func (c *zKFrameworkImpl) watchEvents() {
	log.Printf("watching events from Zookeeper server at %s", c.url)

	c.shutdownConsumers.Add(1)
	defer func() {
		c.shutdownConsumers.Add(-1)
	}()

	for {
		select {
		case <-c.shutdown:
			return
		case event := <-c.events:
			for i := 0; i < int(c.statusChangeConsumers.Load()); i++ {
				c.statusChange <- event.State
			}
		}
	}
}

func (c *zKFrameworkImpl) connectionWatcher() {
	log.Printf("watching connection to Zookeeper server at %s", c.url)

	c.shutdownConsumers.Add(1)
	defer func() {
		c.shutdownConsumers.Add(-1)
	}()

	c.statusChangeConsumers.Add(1)
	defer func() {
		c.statusChangeConsumers.Add(-1)
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

func (c *zKFrameworkImpl) handleStatusChange(state zk.State) {
	c.statusChangeLock.Lock()
	defer c.statusChangeLock.Unlock()

	if state == c.state {
		return
	}

	c.previousState = c.state
	c.state = state
	c.NotifyStatusChange()
	log.Printf("status change from %s to %s", c.previousState, c.state)

	if !c.previouslyConnected() && isConnectedState(c.state) {
		c.reconnectionTimeoutMs = defaultReconnectionTimeoutMs
	}
	if c.started && c.previouslyConnected() && !isConnectedState(c.state) {
		log.Printf("connection to Zookeeper server at %s lost, trying to reconnect", c.url)
		c.invalidateCn()
	}
}

func (c *zKFrameworkImpl) tryConnect() error {
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

func (c *zKFrameworkImpl) invalidateCn() {
	c.stopBgTasks()
	<-time.After(time.Duration(c.reconnectionTimeoutMs) * time.Millisecond)
	c.reconnectionTimeoutMs *= 2

	if c.cn != nil {
		c.cn.Close()
	}
	c.tryConnect()
}

func (c *zKFrameworkImpl) previouslyConnected() bool {
	return isConnectedState(c.previousState)
}

func (c *zKFrameworkImpl) stopBgTasks() {
	for i := 0; i < int(c.shutdownConsumers.Load()); i++ {
		c.shutdown <- true
	}
}

func isConnectedState(state zk.State) bool {
	return state == zk.StateConnected ||
		state == zk.StateHasSession ||
		state == zk.StateConnectedReadOnly ||
		state == zk.StateSaslAuthenticated ||
		state == zk.StateSyncConnected
}

/*
CreateFramework creates a new Zookeeper client with the given connection URL and namespace.
*/
func CreateFramework(url string, namespace ...string) (ZKFramework, error) {
	if url == "" {
		return nil, ErrInvalidConnectionURL
	}

	useNamespace := "/" + strings.TrimPrefix(path.Join(namespace...), "/")

	return &zKFrameworkImpl{
		// TODO more connection options
		namespace: useNamespace,
		url:       url,
		state:     zk.StateDisconnected,
		started:   false,

		shutdownConsumers:     atomic.Int32{},
		statusChangeConsumers: atomic.Int32{},
		reconnectionTimeoutMs: defaultReconnectionTimeoutMs,

		shutdown:              make(chan bool),
		shutdownListeners:     make(map[string]listener.ShutdownListener),
		statusChange:          make(chan zk.State),
		statusChangeListeners: make(map[string]listener.StatusChangeListener),
		statusChangeLock:      sync.RWMutex{},
	}, nil
}
