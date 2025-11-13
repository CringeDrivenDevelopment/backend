package utils

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNotEnoughPerms  = errors.New("not enough permissions")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidInitData = errors.New("invalid init data")
)

func Convert(functionError error) error {
	if errors.Is(functionError, pgx.ErrNoRows) {
		return huma.Error404NotFound("entry not found")
	}

	if errors.Is(functionError, ErrNotEnoughPerms) {
		return huma.Error403Forbidden("not enough permissions")
	}

	if errors.Is(functionError, ErrInvalidToken) {
		return huma.Error401Unauthorized("invalid token")
	}

	if errors.Is(functionError, ErrInvalidInitData) {
		return huma.Error401Unauthorized("invalid init data")
	}

	return huma.Error500InternalServerError("internal server error")
}
