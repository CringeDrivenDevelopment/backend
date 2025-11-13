package handlers

import (
	"backend/internal/app"
	"backend/internal/errorz"
	"backend/internal/service"
	"backend/internal/transport/api/dto"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type User struct {
	userService  *service.User
	tokenService *service.Auth

	logger *zap.Logger
}

// NewAuth - создать новый экземпляр обработчика
func NewAuth(app *app.App) *User {
	userService := service.NewUserService(app)
	tokenService := service.NewAuthService(app, time.Hour)

	return &User{
		userService:  userService,
		tokenService: tokenService,
		logger:       app.Logger,
	}
}

// login - Получить токен для взаимодействия. Нуждается в Raw строке из Telegram Mini App. Действует 1 час
func (h *User) login(ctx context.Context, input *dto.AuthInputStruct) (*dto.AuthOutputStruct, error) {
	id, err := h.tokenService.ParseInitData(input.Body.Raw)
	if err != nil {
		h.logger.Warn("auth data error", zap.Error(err))

		return nil, errorz.Convert(err)
	}

	if err := h.userService.GetByID(ctx, id); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			h.logger.Error("error fetching user", zap.Error(err))

			return nil, errorz.Convert(err)
		}

		if err := h.userService.Create(ctx, id); err != nil {
			h.logger.Error("error creating user", zap.Error(err))

			return nil, errorz.Convert(err)
		}
	}

	token, err := h.tokenService.GenerateToken(id)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))

		return nil, errorz.Convert(err)
	}

	tokenData := dto.Token{
		Token: token,
	}

	return &dto.AuthOutputStruct{Body: tokenData}, nil
}
