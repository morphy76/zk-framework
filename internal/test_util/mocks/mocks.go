package mocks

import (
	"github.com/go-zookeeper/zk"
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
func (m *MockedStatusChangeListener) OnStatusChange(previous zk.State, current zk.State) error {
	m.Interactions++
	return nil
}
