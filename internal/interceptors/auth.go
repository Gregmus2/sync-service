package interceptors

import (
	"context"
	"firebase.google.com/go/auth"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const AuthInterceptorName = "AuthInterceptor"
const ContextFirebaseID = "firebase-id"
const authorizationHeaderName = "authorization"

type AuthInterceptor struct {
	client *auth.Client
}

func (i AuthInterceptor) GetConstructor() any {
	return func(client *auth.Client) (*AuthInterceptor, error) {
		return &AuthInterceptor{client: client}, nil
	}
}

func (i AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, err := i.auth(ctx)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (i AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := i.auth(ss.Context())
		if err != nil {
			return err
		}

		wrapped := middleware.WrapServerStream(ss)
		wrapped.WrappedContext = ctx

		return handler(srv, wrapped)
	}
}

func (i AuthInterceptor) DependsOn() []string {
	return []string{}
}

func (i AuthInterceptor) Name() string {
	return AuthInterceptorName
}

func (i AuthInterceptor) auth(ctx context.Context) (context.Context, error) {
	token := metadata.ExtractIncoming(ctx).Get(authorizationHeaderName)
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "authorization token is required")
	}

	t, err := i.client.VerifyIDTokenAndCheckRevoked(context.Background(), token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return context.WithValue(ctx, ContextFirebaseID, t.UID), nil
}
