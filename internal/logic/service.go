package logic

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/adapters"
	"github.com/pkg/errors"
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
	groupID, err := s.repo.GetGroupID(userID)
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

func (s *service) JoinGroup(userID, groupID string, mergeData bool) ([]*proto.Operation, error) {
	s.mx.Lock(groupID)
	defer s.mx.Unlock(groupID)

	// retrieve all operations from the group first, because later they will be mixed with the user's operations
	operations, err := s.repo.GetAllData(groupID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all data")
	}

	err = s.repo.UpdateGroupID(userID, groupID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update group id")
	}

	if mergeData {
		err = s.repo.MigrateData(userID, groupID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to migrate data")
		}
	} else {
		err = s.repo.RemoveData(userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to remove data")
		}
	}

	return operations, nil
}

func (s *service) LeaveGroup(userID, groupID string, copyData bool) error {
	s.mx.Lock(groupID)
	defer s.mx.Unlock(groupID)

	if copyData {
		err := s.repo.CopyOperations(groupID, userID)
		if err != nil {
			return errors.Wrap(err, "failed to copy operations")
		}
	}

	// todo check if group has any users left and remove group and data if not

	err := s.repo.UpdateGroupID(userID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to update group id")
	}

	return nil
}
