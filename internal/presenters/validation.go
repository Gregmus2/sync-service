package presenters

import (
	"github.com/GregmusCo/poll-play-golang-core/interceptors"
	validation "github.com/go-ozzo/ozzo-validation"
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

	}

	return err
}
