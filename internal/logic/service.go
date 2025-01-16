package logic

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/adapters"
	"github.com/pkg/errors"
)

var (
	ErrGroupNotFound = errors.New("group not found")
	ErrNotInGroup    = errors.New("not in group")
)

type service struct {
	mx GroupMutex

	repo adapters.Repository
	wp   WorkerPool
}

func NewService(mx GroupMutex, repo adapters.Repository, wp WorkerPool) Service {
	return &service{
		mx:   mx,
		repo: repo,
		wp:   wp,
	}
}

func (s *service) SyncData(deviceToken, userID string, stream proto.SyncService_SyncDataServer) error {
	groupID, err := s.repo.GetGroupID(deviceToken, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get group id")
	}

	s.mx.Lock(groupID)
	defer s.mx.Unlock(groupID)

	wg := s.wp.Add(stream)

	data, err := s.repo.GetData(deviceToken, groupID)
	if err != nil {
		return errors.Wrap(err, "failed to get data")
	}

	for _, operation := range data {
		err = stream.Send(operation)
		if err != nil {
			return errors.Wrap(err, "failed to send data")
		}
	}

	wg.Wait()

	if err := s.repo.CleanConflicted(deviceToken, groupID); err != nil {
		return errors.Wrap(err, "failed to clean conflicts")
	}

	if err := s.repo.UpdateDeviceTokenTime(deviceToken, userID, groupID); err != nil {
		return errors.Wrap(err, "failed to update device token time")
	}

	return nil
}

func (s *service) JoinGroup(deviceToken, userID, groupID string, mergeData bool, stream proto.SyncService_JoinGroupServer) error {
	exists, err := s.repo.IsGroupExists(groupID)
	if err != nil {
		return errors.Wrap(err, "failed to check if group exists")
	}
	if !exists {
		return ErrGroupNotFound
	}

	currentGroupID, err := s.repo.GetGroupID(deviceToken, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get group id")
	}

	s.mx.Lock(groupID)
	s.mx.Lock(currentGroupID)
	defer s.mx.Unlock(groupID)
	defer s.mx.Unlock(currentGroupID)

	// retrieve all operations from the group first, because later they will be mixed with the user's operations
	operations, err := s.repo.GetAllData(groupID)
	if err != nil {
		return errors.Wrap(err, "failed to get all data")
	}

	if mergeData {
		unsyncedOperations, err := s.repo.GetData(deviceToken, currentGroupID)
		if err != nil {
			return errors.Wrap(err, "failed to get data")
		}

		operations = append(operations, unsyncedOperations...)
	}

	for _, operation := range operations {
		err = stream.Send(operation)
		if err != nil {
			return errors.Wrap(err, "failed to send data")
		}
	}

	err = s.repo.UpdateGroupID(userID, groupID)
	if err != nil {
		return errors.Wrap(err, "failed to update group id")
	}

	err = s.repo.UpdateDeviceTokenTime(deviceToken, userID, groupID)
	if err != nil {
		return errors.Wrap(err, "failed to update device token time")
	}

	if mergeData {
		err = s.repo.MigrateData(currentGroupID, groupID)
		if err != nil {
			return errors.Wrap(err, "failed to migrate data")
		}
	} else {
		err = s.repo.RemoveData(currentGroupID)
		if err != nil {
			return errors.Wrap(err, "failed to remove data")
		}
	}

	return nil
}

func (s *service) LeaveGroup(deviceToken, userID string, copyData bool) error {
	groupID, err := s.repo.GetGroupID(deviceToken, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get group id")
	}
	if groupID == userID {
		return ErrNotInGroup
	}

	s.mx.Lock(groupID)
	defer s.mx.Unlock(groupID)

	if copyData {
		err := s.repo.CopyOperations(groupID, userID)
		if err != nil {
			return errors.Wrap(err, "failed to copy operations")
		}
	}

	// todo check if group has any users left and remove group and data if not

	err = s.repo.UpdateGroupID(userID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to update group id")
	}

	return nil
}
