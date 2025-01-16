package presenters

import (
	"context"
	sync_proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/adapters"
	"github.com/Gregmus2/sync-service/internal/interceptors"
	"github.com/Gregmus2/sync-service/internal/logic"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Public struct {
	sync_proto.UnimplementedSyncServiceServer

	service logic.Service
	repo    adapters.Repository
}

func NewAPI(service logic.Service, repo adapters.Repository) sync_proto.SyncServiceServer {
	return &Public{
		service: service,
		repo:    repo,
	}
}

func (p Public) SyncData(stream sync_proto.SyncService_SyncDataServer) error {
	deviceToken := stream.Context().Value(interceptors.ContextDeviceToken).(string)
	firebaseID := stream.Context().Value(interceptors.ContextFirebaseID).(string)

	err := p.service.SyncData(deviceToken, firebaseID, stream)
	if err != nil {
		return errors.Wrap(err, "failed to sync data")
	}

	return nil
}

func (p Public) JoinGroup(request *sync_proto.JoinGroupRequest, stream sync_proto.SyncService_JoinGroupServer) error {
	deviceToken := stream.Context().Value(interceptors.ContextDeviceToken).(string)
	firebaseID := stream.Context().Value(interceptors.ContextFirebaseID).(string)

	err := p.service.JoinGroup(deviceToken, firebaseID, request.Group, request.MergeData, stream)
	if err != nil {
		return errors.Wrap(err, "failed to join group")
	}

	return nil
}

func (p Public) LeaveGroup(ctx context.Context, request *sync_proto.LeaveGroupRequest) (*emptypb.Empty, error) {
	deviceToken := ctx.Value(interceptors.ContextDeviceToken).(string)
	firebaseID := ctx.Value(interceptors.ContextFirebaseID).(string)

	err := p.service.LeaveGroup(deviceToken, firebaseID, request.CopyData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to leave group")
	}

	return &emptypb.Empty{}, nil
}

func (p Public) GetCurrentGroup(ctx context.Context, _ *emptypb.Empty) (*sync_proto.GetCurrentGroupResponse, error) {
	deviceToken := ctx.Value(interceptors.ContextDeviceToken).(string)
	firebaseID := ctx.Value(interceptors.ContextFirebaseID).(string)

	groupID, err := p.repo.GetGroupID(deviceToken, firebaseID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current group")
	}

	return &sync_proto.GetCurrentGroupResponse{Group: groupID}, nil
}
