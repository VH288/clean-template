package validator

import (
	"fmt"
	"strings"

	apperrors "clean-template/internal/pkg/errors"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Struct(s any) error {
	if err := validate.Struct(s); err != nil {
		details := make(map[string]string)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				details[strings.ToLower(fieldErr.Field())] = fmt.Sprintf(
					"failed on '%s' validation",
					fieldErr.Tag(),
				)
			}
		}
		return apperrors.WithDetails(apperrors.ErrValidation, details)
	}
	return nil
}
