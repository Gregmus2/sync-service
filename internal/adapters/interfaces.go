package adapters

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
)

type Repository interface {
	UpdateDeviceTokenTime(deviceToken, userID, groupID string) error
	InsertData(deviceToken, groupID string, operation []*proto.Operation) error
	CleanConflicted(deviceToken, groupID string) error
	GetGroupID(deviceToken, userID string) (string, error)
	GetData(deviceToken, groupID string) ([]*proto.SimpleOperation, error)
	UpdateGroupID(userID, newGroupID string) error
	MigrateData(fromID, toID string) error
	RemoveData(groupID string) error
	GetAllData(groupID string) ([]*proto.SimpleOperation, error)
	CopyOperations(fromID, toID string) error
	IsGroupExists(groupID string) (bool, error)
}
