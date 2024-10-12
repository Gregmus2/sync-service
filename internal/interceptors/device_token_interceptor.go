package interceptors

import (
	"context"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DeviceTokenInterceptorName = "DeviceTokenInterceptor"
const deviceTokenHeaderName = "device-token"
const ContextDeviceToken = "device-token"

type DeviceTokenInterceptor struct {
}

func (i DeviceTokenInterceptor) GetConstructor() any {
	return func() (*DeviceTokenInterceptor, error) {
		return &DeviceTokenInterceptor{}, nil
	}
}

func (i DeviceTokenInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, err := i.checkDeviceToken(ctx)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (i DeviceTokenInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := i.checkDeviceToken(ss.Context())
		if err != nil {
			return err
		}

		wrapped := middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}

func (i DeviceTokenInterceptor) DependsOn() []string {
	return []string{}
}

func (i DeviceTokenInterceptor) Name() string {
	return DeviceTokenInterceptorName
}

func (i DeviceTokenInterceptor) checkDeviceToken(ctx context.Context) (context.Context, error) {
	token := metadata.ExtractIncoming(ctx).Get(deviceTokenHeaderName)
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "device token is required")
	}

	return context.WithValue(ctx, ContextDeviceToken, token), nil
}
