package presenters

import (
	"context"
	sync_proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/interceptors"
	"github.com/Gregmus2/sync-service/internal/logic"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Public struct {
	sync_proto.UnimplementedSyncServiceServer

	service logic.Service
}

func NewAPI(service logic.Service) sync_proto.SyncServiceServer {
	return &Public{
		service: service,
	}
}

func (p Public) SyncData(ctx context.Context, request *sync_proto.SyncDataRequest) (*sync_proto.SyncDataResponse, error) {
	deviceToken := ctx.Value(interceptors.ContextDeviceToken).(string)
	firebaseID := ctx.Value(interceptors.ContextFirebaseID).(string)

	operations, err := p.service.SyncData(deviceToken, firebaseID, request.Operations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sync data")
	}

	return &sync_proto.SyncDataResponse{Operations: operations}, nil
}

func (p Public) JoinGroup(ctx context.Context, request *sync_proto.JoinGroupRequest) (*emptypb.Empty, error) {

}

func (p Public) LeaveGroup(ctx context.Context, request *sync_proto.LeaveGroupRequest) (*emptypb.Empty, error) {

}
