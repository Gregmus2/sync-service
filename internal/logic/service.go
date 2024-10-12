package logic

import (
	proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/adapters"
	"github.com/pkg/errors"
)

type service struct {
	mx GroupMutex

	repo adapters.Repository
}

func NewService(mx GroupMutex, repo adapters.Repository) Service {
	return &service{
		mx:   mx,
		repo: repo,
	}
}

func (s *service) SyncData(deviceToken, userID string, operations []*proto.Operation) ([]*proto.Operation, error) {
	groupID, err := s.repo.GetGroupID(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get group id")
	}

	s.mx.Lock(groupID)
	defer s.mx.Unlock(groupID)

	if err := s.repo.InsertData(deviceToken, groupID, operations); err != nil {
		return nil, errors.Wrap(err, "failed to insert data")
	}

	if err := s.repo.CleanConflicted(deviceToken, groupID); err != nil {
		return nil, errors.Wrap(err, "failed to remove conflicts")
	}

	data, err := s.repo.GetData(deviceToken, groupID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get data")
	}

	err = s.repo.UpdateDeviceTokenTime(deviceToken, userID, groupID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sync device token")
	}

	return data, nil
}
