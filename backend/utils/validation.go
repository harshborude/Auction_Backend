package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func FormatValidationErrors(err error) map[string]string {

	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {

		for _, fieldErr := range validationErrors {

			field := fieldErr.Field()

			switch fieldErr.Tag() {

			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)

			case "email":
				errors[field] = "invalid email format"

			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s characters", field, fieldErr.Param())

			case "max":
				errors[field] = fmt.Sprintf("%s must be at most %s characters", field, fieldErr.Param())

			case "alphanum":
				errors[field] = fmt.Sprintf("%s must contain only letters and numbers", field)

			default:
				errors[field] = fmt.Sprintf("%s is invalid", field)

			}
		}
	}

	return errors
}