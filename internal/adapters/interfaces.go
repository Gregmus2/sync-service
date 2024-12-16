package adapters

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
)

type Repository interface {
	UpdateDeviceTokenTime(deviceToken, userID, groupID string) error
	InsertData(deviceToken, groupID string, operations []*proto.Operation) error
	CleanConflicted(deviceToken, groupID string) error
	GetGroupID(userID string) (string, error)
	GetData(deviceToken, groupID string) ([]*proto.Operation, error)
	UpdateGroupID(userID, newGroupID string) error
	MigrateData(fromID, toID string) error
	RemoveData(groupID string) error
	GetAllData(groupID string) ([]*proto.Operation, error)
	CopyOperations(fromID, toID string) error
}
