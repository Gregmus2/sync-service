package logic

import proto "github.com/Gregmus2/sync-proto-gen/go/sync"

type Service interface {
	SyncData(deviceToken, userID string, operations []*proto.Operation) ([]*proto.Operation, error)
}

type GroupMutex interface {
	Lock(groupID string)
	Unlock(groupID string)
}