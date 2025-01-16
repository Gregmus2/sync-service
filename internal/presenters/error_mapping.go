package presenters

import (
	"github.com/Gregmus2/go-grpc-core/interceptors"
	"github.com/Gregmus2/sync-service/internal/logic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewErrorMapping() interceptors.ErrorMapping {
	return interceptors.ErrorMapping{
		logic.ErrGroupNotFound: status.Error(codes.NotFound, "group not found"),
		logic.ErrNotInGroup:    status.Error(codes.InvalidArgument, "you can't leave own group"),
	}
}
