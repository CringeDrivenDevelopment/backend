package handlers

import (
	"backend/internal/interfaces"
	"backend/internal/service"
	"backend/internal/transport/api/dto"
	"backend/pkg/utils"
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Auth struct {
	userService interfaces.UserService
	authService interfaces.AuthService

	logger *zap.Logger
}

// NewAuth - создать новый экземпляр обработчика
func NewAuth(userService *service.User, authService *service.Auth, logger *zap.Logger, api huma.API) *Auth {
	result := &Auth{
		userService: userService,
		authService: authService,
		logger:      logger,
	}

	result.setup(api)

	return result
}

// login - Получить токен для взаимодействия. Нуждается в Raw строке из Telegram Mini App. Действует 1 час
func (h *Auth) login(ctx context.Context, input *dto.AuthInputStruct) (*dto.AuthOutputStruct, error) {
	id, err := h.authService.ParseInitData(input.Body.Raw)
	if err != nil {
		h.logger.Warn("login error", zap.Error(err))

		return nil, utils.Convert(err)
	}

	if err := h.userService.GetByID(ctx, id); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			h.logger.Error("login error", zap.Error(err))

			return nil, utils.Convert(err)
		}

		if err := h.userService.Create(ctx, id); err != nil {
			h.logger.Error("login error", zap.Error(err))

			return nil, utils.Convert(err)
		}
	}

	token, err := h.authService.GenerateToken(id)
	if err != nil {
		h.logger.Error("login error", zap.Error(err))

		return nil, utils.Convert(err)
	}

	tokenData := dto.Token{
		Token: token,
	}

	return &dto.AuthOutputStruct{Body: tokenData}, nil
}
