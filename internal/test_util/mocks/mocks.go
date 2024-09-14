/*
Package mocks provides mocked implementations of various interfaces for use in unit tests.
*/
package mocks

import (
	"github.com/go-zookeeper/zk"
	"github.com/morphy76/zk/pkg/core"
)

/*
MockedStatusChangeListener is a mocked implementation of the StatusChangeListener interface.
*/
type MockedStatusChangeListener struct {
	ID           string
	Interactions uint
}

/*
UUID is a mocked implementation of the UUID method.
*/
func (m *MockedStatusChangeListener) UUID() string {
	return m.ID
}

/*
OnStatusChange is a mocked implementation of the OnStatusChange method.
*/
func (m *MockedStatusChangeListener) OnStatusChange(zkFramework core.ZKFramework, previous zk.State, current zk.State) error {
	m.Interactions++
	return nil
}

/*
Stop is a mocked implementation of the Stop method.
*/
func (m *MockedStatusChangeListener) Stop() {
}

/*
MockedShutdownListener is a mocked implementation of the ShutdownListener interface.
*/
type MockedShutdownListener struct {
	ID           string
	Interactions uint
}

/*
UUID is a mocked implementation of the UUID method.
*/
func (m *MockedShutdownListener) UUID() string {
	return m.ID
}

/*
OnShutdown is a mocked implementation of the OnShutdown method.
*/
func (m *MockedShutdownListener) OnShutdown(zkFramework core.ZKFramework) error {
	m.Interactions++
	return nil
}

/*
Stop is a mocked implementation of the Stop method.
*/
func (m *MockedShutdownListener) Stop() {
}
