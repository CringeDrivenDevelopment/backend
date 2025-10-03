package handlers

import (
	"backend/internal/application"
	"backend/internal/domain/models"
	"backend/internal/domain/service"
	"context"
	"errors"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type User struct {
	userService  *service.User
	tokenService *service.Auth

	logger *zap.Logger
}

// NewAuth - создать новый экземпляр обработчика
func NewAuth(app *application.App) *User {
	userService := service.NewUserService(app)
	tokenService := service.NewAuthService(app, time.Hour)

	return &User{
		userService:  userService,
		tokenService: tokenService,
		logger:       app.Logger,
	}
}

// login - Получить токен для взаимодействия. Нуждается в InitDataRaw строке из Telegram Mini App. Действует 1 час
func (h *User) login(ctx context.Context, input *models.AuthInputStruct) (*models.AuthOutputStruct, error) {
	userDTO := input.Body

	id, tgErr := h.tokenService.ParseInitData(userDTO.InitDataRaw)
	if tgErr != nil {
		h.logger.Warn("auth data error", zap.Error(tgErr))

		return nil, huma.Error401Unauthorized("invalid auth data")
	}

	errFetch := h.userService.GetByID(ctx, id)
	if errFetch != nil {
		if !errors.Is(errFetch, pgx.ErrNoRows) {
			h.logger.Error("error fetching user", zap.Error(errFetch))

			return nil, huma.Error500InternalServerError("internal server error")
		}

		createErr := h.userService.Create(ctx, id)

		if createErr != nil {
			h.logger.Error("error creating user", zap.Error(createErr))

			return nil, huma.Error500InternalServerError("internal server error")
		}
	}

	token, tokenErr := h.tokenService.GenerateToken(id)
	if tokenErr != nil {
		h.logger.Error("failed to generate token", zap.Error(tokenErr))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	tokenData := models.TokenDto{
		Token: token,
	}

	return &models.AuthOutputStruct{Body: tokenData}, nil
}
