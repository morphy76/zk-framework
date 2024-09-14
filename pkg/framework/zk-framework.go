/*
Package framework provides a higher level Zookeeper client with more capabilities than the standard Zookeeper client.
*/
package framework

import (
	"log"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/core"
	"github.com/morphy76/zk/pkg/core/coreerr"
	"github.com/morphy76/zk/pkg/framework/frwkerr"
)

const (
	defaultReconnectionTimeoutMs = 100
)

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
	shutdownListeners map[string]core.ShutdownListener

	statusChange          chan zk.State
	statusChangeConsumers atomic.Int32
	statusChangeLock      sync.RWMutex
	statusChangeListeners map[string]core.StatusChangeListener
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
		return frwkerr.ErrFrameworkAlreadyStarted
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
		return frwkerr.ErrFrameworkNotYetStarted
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
			return frwkerr.ErrConnectionTimeout
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
		return frwkerr.ErrFrameworkNotYetStarted
	}
	defer c.cn.Close()

	log.Printf("closing connection to Zookeeper server at %s", c.url)

	c.stopBgTasks()
	go func() {
		c.NotifyShutdown()
		c.clearAllListeners()
	}()

	c.started = false
	c.state = zk.StateDisconnected

	return nil
}

/*
AddStatusChangeListener adds a listener for Zookeeper connection status changes.
*/
func (c *zKFrameworkImpl) AddStatusChangeListener(statusChangeListener core.StatusChangeListener) error {
	// TODO locks

	if found := c.statusChangeListeners[statusChangeListener.UUID()]; found != nil {
		return coreerr.ErrListenerAlreadyExists
	}

	c.statusChangeListeners[statusChangeListener.UUID()] = statusChangeListener
	return nil
}

/*
RemoveStatusChangeListener removes a listener for Zookeeper connection status changes.
*/
func (c *zKFrameworkImpl) RemoveStatusChangeListener(statusChangeListener core.StatusChangeListener) error {
	// TODO locks

	if found := c.statusChangeListeners[statusChangeListener.UUID()]; found == nil {
		return coreerr.ErrListenerNotFound
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
		if err := listener.OnStatusChange(c, c.previousState, c.state); err != nil {
			log.Printf("error notifying status change listener: %s", err)
		}
	}
}

/*
AddShutdownListener adds a listener for Zookeeper client shutdown events.
*/
func (c *zKFrameworkImpl) AddShutdownListener(shutdownListener core.ShutdownListener) error {
	if found := c.shutdownListeners[shutdownListener.UUID()]; found != nil {
		return coreerr.ErrListenerAlreadyExists
	}

	c.shutdownListeners[shutdownListener.UUID()] = shutdownListener
	return nil
}

/*
RemoveShutdownListener removes a listener for Zookeeper client shutdown events.
*/
func (c *zKFrameworkImpl) RemoveShutdownListener(shutdownListener core.ShutdownListener) error {
	if found := c.shutdownListeners[shutdownListener.UUID()]; found == nil {
		return coreerr.ErrListenerNotFound
	}

	delete(c.shutdownListeners, shutdownListener.UUID())
	return nil
}

/*
NotifyShutdown notifies all listeners of a Zookeeper client shutdown event.
*/
func (c *zKFrameworkImpl) NotifyShutdown() {
	for _, listener := range c.shutdownListeners {
		if err := listener.OnShutdown(c); err != nil {
			log.Printf("error notifying shutdown listener: %s", err)
		}
	}
}

func (c *zKFrameworkImpl) clearAllListeners() {
	for _, listener := range c.statusChangeListeners {
		listener.Stop()
	}
	c.statusChangeListeners = make(map[string]core.StatusChangeListener)

	for _, listener := range c.shutdownListeners {
		listener.Stop()
	}
	c.shutdownListeners = make(map[string]core.ShutdownListener)
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
	go c.NotifyStatusChange()
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
func CreateFramework(url string, namespace ...string) (core.ZKFramework, error) {
	if url == "" {
		return nil, frwkerr.ErrInvalidConnectionURL
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
		shutdownListeners:     make(map[string]core.ShutdownListener),
		statusChange:          make(chan zk.State),
		statusChangeListeners: make(map[string]core.StatusChangeListener),
		statusChangeLock:      sync.RWMutex{},
	}, nil
}
