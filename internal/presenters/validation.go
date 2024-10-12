package presenters

import (
	"github.com/Gregmus2/go-grpc-core/interceptors"
	sync_proto "github.com/Gregmus2/sync-proto-gen/go/sync"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type validator struct{}

func NewValidator() interceptors.Validator {
	return &validator{}
}

func (v validator) Validate(request any) error {
	err := validation.Validate(request, validation.Required)
	if err != nil {
		return err
	}

	switch r := request.(type) {
	case *sync_proto.SyncDataRequest:
		for _, op := range r.Operations {
			return validation.ValidateStruct(op,
				validation.Field(&op.Type, validation.Required),
				validation.Field(&op.Sql, validation.Required),
				validation.Field(&op.Id, validation.Required, is.UUIDv4),
				validation.Field(&op.Entity, validation.Required),
			)
		}
	}

	return nil
}
