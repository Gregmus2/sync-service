package mocks

import "github.com/stretchr/testify/mock"

// MockGroupMutex is a mock of the GroupMutex interface
type MockGroupMutex struct {
	mock.Mock
}

func (m *MockGroupMutex) Lock(groupID string) {
	m.Called(groupID)
}

func (m *MockGroupMutex) Unlock(groupID string) {
	m.Called(groupID)
}
