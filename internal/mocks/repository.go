package mocks

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) UpdateDeviceTokenTime(deviceToken, userID, groupID string) error {
	args := m.Called(deviceToken, userID, groupID)
	return args.Error(0)
}

func (m *MockRepository) InsertData(deviceToken, groupID string, operations []*proto.Operation) error {
	args := m.Called(deviceToken, groupID, operations)
	return args.Error(0)
}

func (m *MockRepository) CleanConflicted(deviceToken, groupID string) error {
	args := m.Called(deviceToken, groupID)
	return args.Error(0)
}

func (m *MockRepository) GetGroupID(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetData(deviceToken, groupID string) ([]*proto.Operation, error) {
	args := m.Called(deviceToken, groupID)
	return args.Get(0).([]*proto.Operation), args.Error(1)
}

func (m *MockRepository) UpdateGroupID(userID, newGroupID string) error {
	args := m.Called(userID, newGroupID)
	return args.Error(0)
}

func (m *MockRepository) MigrateData(fromID, toID string) error {
	args := m.Called(fromID, toID)
	return args.Error(0)
}

func (m *MockRepository) RemoveData(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockRepository) GetAllData(groupID string) ([]*proto.Operation, error) {
	args := m.Called(groupID)
	return args.Get(0).([]*proto.Operation), args.Error(1)
}

func (m *MockRepository) CopyOperations(fromID, toID string) error {
	args := m.Called(fromID, toID)
	return args.Error(0)
}
