package validator

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

type Validator struct {
	validator *validator.Validate
}

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

func New(botTokens []string) *Validator {
	newValidator := validator.New()
	// TODO: change to a faster validation engine

	_ = newValidator.RegisterValidation("init_data_raw", func(fl validator.FieldLevel) bool {
		initDataRaw := fl.Field().String()
		expIn := 1 * time.Hour

		for _, token := range botTokens {
			if initdata.Validate(initDataRaw, token, expIn) == nil {
				return true
			}
		}

		return false
	})

	return &Validator{
		newValidator,
	}
}

func (v *Validator) ValidateData(data interface{}) error {
	var validationErrors []ErrorResponse

	errs := v.validator.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = err.Field() // Export struct field name
			elem.Tag = err.Tag()           // Export struct tag
			elem.Value = err.Value()       // Export field value
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	if len(validationErrors) > 0 && validationErrors[0].Error {
		errMessages := make([]string, 0)

		for _, err := range validationErrors {
			errMessages = append(errMessages, fmt.Sprintf(
				"[%s]: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}

		return errors.New(strings.Join(errMessages, " and "))
	}
	return nil
}
