package adapters

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
)

type Repository interface {
	UpdateDeviceTokenTime(deviceToken, userID, groupID string) error
	InsertData(deviceToken, groupID string, operations []*proto.Operation) error
	// todo Check only conflicts with the same groupID and from the last sync of this device
	CleanConflicted(deviceToken, groupID string) error
	GetGroupID(userID string) (string, error)
	GetData(deviceToken, groupID string) ([]*proto.Operation, error)
}
