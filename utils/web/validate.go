package web

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func GetValidator() *validator.Validate {
	if validate == nil {
		validate = validator.New(validator.WithRequiredStructEnabled())
	}
	return validate
}

func GetValidationError(err error) string {
	return err.(validator.ValidationErrors)[0].Error()
}
