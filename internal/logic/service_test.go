package logic_test

import (
	"errors"
	"github.com/Gregmus2/sync-service/internal/logic"
	"github.com/Gregmus2/sync-service/internal/mocks"
	"testing"

	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mocks.MockRepository is a mock of the Repository interface

func TestService_SyncData(t *testing.T) {
	t.Run("successful sync", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockRepo.On("GetGroupID", "user1").Return("group1", nil)
		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("InsertData", "device1", "group1", mock.Anything).Return(nil)
		mockRepo.On("CleanConflicted", "device1", "group1").Return(nil)
		mockRepo.On("GetData", "device1", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateDeviceTokenTime", "device1", "user1", "group1").Return(nil)
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.SyncData("device1", "user1", []*proto.Operation{})
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("GetGroupID error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockRepo.On("GetGroupID", "user1").Return("", errors.New("get group id error"))

		_, err := service.SyncData("device1", "user1", []*proto.Operation{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get group id error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("InsertData error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockRepo.On("GetGroupID", "user1").Return("group1", nil)
		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("InsertData", "device1", "group1", mock.Anything).Return(errors.New("insert data error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.SyncData("device1", "user1", []*proto.Operation{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insert data error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("CleanConflicted error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockRepo.On("GetGroupID", "user1").Return("group1", nil)
		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("InsertData", "device1", "group1", mock.Anything).Return(nil)
		mockRepo.On("CleanConflicted", "device1", "group1").Return(errors.New("clean conflicted error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.SyncData("device1", "user1", []*proto.Operation{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clean conflicted error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("GetData error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockRepo.On("GetGroupID", "user1").Return("group1", nil)
		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("InsertData", "device1", "group1", mock.Anything).Return(nil)
		mockRepo.On("CleanConflicted", "device1", "group1").Return(nil)
		mockRepo.On("GetData", "device1", "group1").Return(([]*proto.Operation)(nil), errors.New("get data error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.SyncData("device1", "user1", []*proto.Operation{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get data error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("UpdateDeviceTokenTime error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockRepo.On("GetGroupID", "user1").Return("group1", nil)
		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("InsertData", "device1", "group1", mock.Anything).Return(nil)
		mockRepo.On("CleanConflicted", "device1", "group1").Return(nil)
		mockRepo.On("GetData", "device1", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateDeviceTokenTime", "device1", "user1", "group1").Return(errors.New("update device token time error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.SyncData("device1", "user1", []*proto.Operation{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update device token time error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})
}

func TestService_JoinGroup(t *testing.T) {
	t.Run("successful join with merge", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("GetAllData", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateGroupID", "user1", "group1").Return(nil)
		mockRepo.On("MigrateData", "user1", "group1").Return(nil)
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.JoinGroup("user1", "group1", true)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("successful join without merge", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("GetAllData", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateGroupID", "user1", "group1").Return(nil)
		mockRepo.On("RemoveData", "user1").Return(nil)
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.JoinGroup("user1", "group1", false)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})
	t.Run("GetAllData error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("GetAllData", "group1").Return(([]*proto.Operation)(nil), errors.New("get all data error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.JoinGroup("user1", "group1", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get all data error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("UpdateGroupID error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("GetAllData", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateGroupID", "user1", "group1").Return(errors.New("update group id error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.JoinGroup("user1", "group1", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update group id error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("MigrateData error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("GetAllData", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateGroupID", "user1", "group1").Return(nil)
		mockRepo.On("MigrateData", "user1", "group1").Return(errors.New("migrate data error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.JoinGroup("user1", "group1", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migrate data error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("RemoveData error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("GetAllData", "group1").Return([]*proto.Operation{}, nil)
		mockRepo.On("UpdateGroupID", "user1", "group1").Return(nil)
		mockRepo.On("RemoveData", "user1").Return(errors.New("remove data error"))
		mockMutex.On("Unlock", "group1").Return()

		_, err := service.JoinGroup("user1", "group1", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "remove data error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})
}

func TestService_LeaveGroup(t *testing.T) {
	t.Run("successful leave with copy", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("CopyOperations", "group1", "user1").Return(nil)
		mockRepo.On("UpdateGroupID", "user1", "user1").Return(nil)
		mockMutex.On("Unlock", "group1").Return()

		err := service.LeaveGroup("user1", "group1", true)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("successful leave without copy", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("UpdateGroupID", "user1", "user1").Return(nil)
		mockMutex.On("Unlock", "group1").Return()

		err := service.LeaveGroup("user1", "group1", false)
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("CopyOperations error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("CopyOperations", "group1", "user1").Return(errors.New("copy operations error"))
		mockMutex.On("Unlock", "group1").Return()

		err := service.LeaveGroup("user1", "group1", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "copy operations error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})

	t.Run("UpdateGroupID error", func(t *testing.T) {
		mockRepo := &mocks.MockRepository{}
		mockMutex := &mocks.MockGroupMutex{}
		service := logic.NewService(mockMutex, mockRepo)

		mockMutex.On("Lock", "group1").Return()
		mockRepo.On("CopyOperations", "group1", "user1").Return(nil)
		mockRepo.On("UpdateGroupID", "user1", "user1").Return(errors.New("update group id error"))
		mockMutex.On("Unlock", "group1").Return()

		err := service.LeaveGroup("user1", "group1", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update group id error")

		mockRepo.AssertExpectations(t)
		mockMutex.AssertExpectations(t)
	})
}
