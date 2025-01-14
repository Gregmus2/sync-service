package logic

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"sync"
)

type Service interface {
	SyncData(deviceToken, userID string, server proto.SyncService_SyncDataServer) error
	JoinGroup(userID, groupID string, mergeData bool) ([]*proto.Operation, error)
	LeaveGroup(userID, groupID string, copyData bool) error
}

type GroupMutex interface {
	Lock(groupID string)
	Unlock(groupID string)
}

type WorkerPool interface {
	Add(server proto.SyncService_SyncDataServer) *sync.WaitGroup
}
