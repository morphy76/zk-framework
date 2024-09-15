package mocks

import (
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/core"
)

/*
SpiedFramework is a spy for the ZKFramework.
*/
type SpiedFramework struct {
	zkFramework  core.ZKFramework
	Interactions map[string]int
}

/*
NewSpiedFramework creates a new SpiedFramework.
*/
func NewSpiedFramework(zkFramework core.ZKFramework) *SpiedFramework {
	return &SpiedFramework{
		zkFramework:  zkFramework,
		Interactions: make(map[string]int),
	}
}

/*
Start starts the Zookeeper client.
*/
func (s *SpiedFramework) Start() error {
	s.Interactions["Start"]++
	return s.zkFramework.Start()
}

/*
Stop stops the Zookeeper client.
*/
func (s *SpiedFramework) Stop() error {
	s.Interactions["Stop"]++
	return s.zkFramework.Stop()
}

/*
AddStatusChangeListener adds a status change listener.
*/
func (s *SpiedFramework) AddStatusChangeListener(listener core.StatusChangeListener) error {
	s.Interactions["AddStatusChangeListener"]++
	return s.zkFramework.AddStatusChangeListener(listener)
}

/*
RemoveStatusChangeListener removes a status change listener.
*/
func (s *SpiedFramework) RemoveStatusChangeListener(listener core.StatusChangeListener) error {
	s.Interactions["RemoveStatusChangeListener"]++
	return s.zkFramework.RemoveStatusChangeListener(listener)
}

/*
NotifyStatusChange notifies a status change.
*/
func (s *SpiedFramework) NotifyStatusChange() {
	s.Interactions["NotifyStatusChange"]++
	s.zkFramework.NotifyStatusChange()
}

/*
AddShutdownListener adds a shutdown listener.
*/
func (s *SpiedFramework) AddShutdownListener(listener core.ShutdownListener) error {
	s.Interactions["AddShutdownListener"]++
	return s.zkFramework.AddShutdownListener(listener)
}

/*
RemoveShutdownListener removes a shutdown listener.
*/
func (s *SpiedFramework) RemoveShutdownListener(listener core.ShutdownListener) error {
	s.Interactions["RemoveShutdownListener"]++
	return s.zkFramework.RemoveShutdownListener(listener)
}

/*
NotifyShutdown notifies a shutdown.
*/
func (s *SpiedFramework) NotifyShutdown() {
	s.Interactions["NotifyShutdown"]++
	s.zkFramework.NotifyShutdown()
}

/*
Namespace gets the namespace.
*/
func (s *SpiedFramework) Namespace() string {
	s.Interactions["Namespace"]++
	return s.zkFramework.Namespace()
}

/*
Cn gets the Zookeeper connection.
*/
func (s *SpiedFramework) Cn() *zk.Conn {
	s.Interactions["Cn"]++
	return s.zkFramework.Cn()
}

/*
URL gets the URL.
*/
func (s *SpiedFramework) URL() string {
	s.Interactions["URL"]++
	return s.zkFramework.URL()
}

/*
Started checks if the Zookeeper client is started.
*/
func (s *SpiedFramework) Started() bool {
	s.Interactions["Started"]++
	return s.zkFramework.Started()
}

/*
Connected checks if the Zookeeper client is connected.
*/
func (s *SpiedFramework) Connected() bool {
	s.Interactions["Connected"]++
	return s.zkFramework.Connected()
}

/*
WaitConnection waits for the connection.
*/
func (s *SpiedFramework) WaitConnection(timeout time.Duration) error {
	s.Interactions["WaitConnection"]++
	return s.zkFramework.WaitConnection(timeout)
}
