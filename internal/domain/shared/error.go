package shared

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

// HandleError - обработать ошибку
func HandleError(err error) *ApiError {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return &ApiError{
			Code:    404,
			Message: "запись не найдена",
		}
	}

	return &ApiError{
		Code:    500,
		Message: "внутренняя ошибка сервера",
	}
}
