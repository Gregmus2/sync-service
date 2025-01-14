package main

import (
	core "github.com/Gregmus2/go-grpc-core"
	"github.com/Gregmus2/go-grpc-core/interceptors"
	sync_proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	"github.com/Gregmus2/sync-service/internal/adapters"
	"github.com/Gregmus2/sync-service/internal/common"
	interceptors2 "github.com/Gregmus2/sync-service/internal/interceptors"
	"github.com/Gregmus2/sync-service/internal/logic"
	"github.com/Gregmus2/sync-service/internal/presenters"
	"go.uber.org/fx"
)

func main() {
	core.Serve(
		[]core.Server{
			{
				Services: []core.Service{
					{ServiceDesc: sync_proto.SyncService_ServiceDesc, Constructor: presenters.NewAPI},
				},
				Interceptors: []interceptors.Interceptor{
					&interceptors.ErrorHandlingInterceptor{},
					&interceptors.RequestValidationInterceptor{},
					&interceptors2.AuthInterceptor{},
					&interceptors2.DeviceTokenInterceptor{},
				},
				Stream: true,
			},
		},
		fx.Provide(
			common.NewConfig,
			adapters.NewDB,
			adapters.NewRepository,
			adapters.NewFirebaseApp,
			adapters.NewFirebaseClient,
			logic.NewGroupMutex,
			logic.NewService,
			logic.NewWorkerPool,
			presenters.NewErrorMapping,
			presenters.NewValidator,
		),
	)
}
