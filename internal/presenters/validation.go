package presenters

import (
	"github.com/Gregmus2/go-grpc-core/interceptors"
)

type validator struct{}

func NewValidator() interceptors.Validator {
	return &validator{}
}

func (v validator) Validate(request any) error {
	return nil
}
