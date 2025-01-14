package presenters

import (
	"context"
	ccvt "github.com/Gregmus2/common-cvt"
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

func (p Public) JoinGroup(request *sync_proto.JoinGroupRequest, server sync_proto.SyncService_JoinGroupServer) error {
	panic("implement me")
	//firebaseID := server.Context().Value(interceptors.ContextFirebaseID).(string)

	//operations, err := p.service.JoinGroup(firebaseID, request.Group, request.MergeData)
	//if err != nil {
	//	return errors.Wrap(err, "failed to join group")
	//}

	return nil
}

func (p Public) LeaveGroup(ctx context.Context, request *sync_proto.LeaveGroupRequest) (*emptypb.Empty, error) {
	firebaseID := ctx.Value(interceptors.ContextFirebaseID).(string)

	err := p.service.LeaveGroup(firebaseID, request.Group, request.CopyData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to leave group")
	}

	return &emptypb.Empty{}, nil
}

func (p Public) GetCurrentGroup(ctx context.Context, request *emptypb.Empty) (*sync_proto.GetCurrentGroupResponse, error) {
	firebaseID := ctx.Value(interceptors.ContextFirebaseID).(string)

	groupID, err := p.repo.GetCurrentGroup(firebaseID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current group")
	}

	return &sync_proto.GetCurrentGroupResponse{Group: ccvt.ToProtoStringWrapper(groupID)}, nil
}
