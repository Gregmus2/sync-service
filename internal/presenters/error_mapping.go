package presenters

import (
	"github.com/Gregmus2/go-grpc-core/interceptors"
)

func NewErrorMapping() interceptors.ErrorMapping {
	return interceptors.ErrorMapping{}
}
