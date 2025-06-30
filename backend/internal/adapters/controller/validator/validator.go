package validator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"strings"
	"time"
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

func New(botToken string) *Validator {
	newValidator := validator.New()

	_ = newValidator.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 4 && len(fl.Field().String()) <= 20
	})

	_ = newValidator.RegisterValidation("init_data_raw", func(fl validator.FieldLevel) bool {
		initDataRaw := fl.Field().String()
		token := botToken
		expIn := 1 * time.Hour

		return initdata.Validate(initDataRaw, token, expIn) == nil
	})

	_ = newValidator.RegisterValidation("header", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 5 && len(fl.Field().String()) <= 150
	})

	_ = newValidator.RegisterValidation("body", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String()) >= 5 && len(fl.Field().String()) <= 1500
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

/*
func (v Validator) GetLimitAndOffset(c fiber.Ctx, defaultLimit string, defaultOffset string) (int, int) {
	limit, err := strconv.Atoi(c.Query("limit", defaultLimit))
	if err != nil {
		return 0, 10
	}
	offset, err := strconv.Atoi(c.Query("offset", defaultOffset))
	if err != nil {
		return 0, 10
	}
	return limit, offset
}
*/
