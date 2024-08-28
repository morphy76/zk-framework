package framework

import (
	"errors"
	"log"
	"time"

	"github.com/go-zookeeper/zk"
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

/*
ZKFrameworkState represents the state of the Zookeeper client.
*/
type ZKFrameworkState int

const (
	Connected         ZKFrameworkState = ZKFrameworkState(zk.StateConnected)
	Disconnected      ZKFrameworkState = ZKFrameworkState(zk.StateDisconnected)
	AuthFailed        ZKFrameworkState = ZKFrameworkState(zk.StateAuthFailed)
	Expired           ZKFrameworkState = ZKFrameworkState(zk.StateExpired)
	Unknown           ZKFrameworkState = ZKFrameworkState(zk.StateUnknown)
	ConnectedReadOnly ZKFrameworkState = ZKFrameworkState(zk.StateConnectedReadOnly)
	Connecting        ZKFrameworkState = ZKFrameworkState(zk.StateConnecting)
	SaslAuthenticated ZKFrameworkState = ZKFrameworkState(zk.StateSaslAuthenticated)
	SyncConnected     ZKFrameworkState = ZKFrameworkState(zk.StateSyncConnected)
)

/*
ZKFramework represents a Zookeeper client with higher level capabilities, wrapping github.com/go-zookeeper/zk.
*/
type ZKFramework struct {
	url           string
	state         ZKFrameworkState
	previousState ZKFrameworkState
	started       bool

	cn     *zk.Conn
	events <-chan zk.Event

	shutdown          chan bool
	shutdownConsumers int

	statusChange          chan zk.State
	statusChangeConsumers int
}

/*
Url returns the URL of the Zookeeper client.
*/
func (c *ZKFramework) Url() string {
	return c.url
}

/*
State returns the state of the Zookeeper client.
*/
func (c *ZKFramework) State() ZKFrameworkState {
	return c.state
}

/*
Started returns whether the Zookeeper client is started.
*/
func (c *ZKFramework) Started() bool {
	return c.started
}

func (c *ZKFramework) Connected() bool {
	return c.state == Connected
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
	cn, events, err := zk.Connect([]string{c.url}, 10*time.Second)

	if err != nil {
		return err
	}

	c.cn = cn
	c.events = events
	c.shutdown = make(chan bool)
	c.statusChange = make(chan zk.State)

	go c.connectionWatcher()
	go c.watchEvents()
	return nil
}

/*
WaitConnection waits for the connection to the Zookeeper server to be established.
*/
func (c *ZKFramework) WaitConnection(timeout time.Duration) error {
	if !c.started {
		return ErrFrameworkNotYetStarted
	}

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
			if ZKFrameworkState(state) == Connected {
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

	log.Printf("closing connection to Zookeeper server at %s", c.url)

	for i := 0; i < c.shutdownConsumers; i++ {
		c.shutdown <- true
	}

	c.cn.Close()
	c.state = Disconnected
	c.started = false

	close(c.shutdown)
	close(c.statusChange)
	c.statusChange = nil

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
				if c.statusChange != nil {
					c.statusChange <- event.State
				}
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
			c.previousState = c.state
			c.state = ZKFrameworkState(state)
			if c.started && c.previousState == Connected && c.state == Disconnected {
				log.Printf("connection to Zookeeper server at %s lost, trying to reconnect", c.url)
				// try to reconnect, see retry policies
				// https://curator.apache.org/apidocs/org/apache/curator/RetryPolicy.html
			}
		}
	}
}

func CreateFramework(url string) (*ZKFramework, error) {
	if url == "" {
		return nil, ErrInvalidConnectionURL
	}

	return &ZKFramework{
		url:     url,
		state:   Disconnected,
		started: false,

		shutdownConsumers:     0,
		statusChangeConsumers: 0,
	}, nil
}
