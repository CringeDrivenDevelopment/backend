package rest

import (
	"backend/internal/application"
	"backend/internal/application/service"
	"backend/internal/domain/models"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type userHandler struct {
	userService  *service.User
	tokenService *service.Auth

	logger *zap.Logger
}

// newUserHandler - создать новый экземпляр обработчика
func newUserHandler(app *application.App) *userHandler {
	userService := service.NewUserService(app)
	tokenService := service.NewAuthService(app, time.Hour)

	return &userHandler{
		userService:  userService,
		tokenService: tokenService,
		logger:       app.Logger,
	}
}

// login - Получить токен для взаимодействия. Нуждается в InitDataRaw строке из Telegram Mini App. Действует 1 час
func (h *userHandler) login(ctx context.Context, input *models.AuthInputStruct) (*models.AuthOutputStruct, error) {
	userDTO := input.Body

	id, tgErr := h.tokenService.ParseInitData(userDTO.InitDataRaw)
	if tgErr != nil {
		return nil, tgErr
	}

	errFetch := h.userService.GetByID(ctx, id)
	if errFetch != nil {
		if !errors.Is(errFetch, pgx.ErrNoRows) {
			h.logger.Warn("error fetching user", zap.Error(errFetch))
			return nil, huma.Error500InternalServerError("internal server error")
		}

		createErr := h.userService.Create(ctx, id)

		if createErr != nil {
			h.logger.Warn("error creating user", zap.Error(createErr))
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

// Setup - добавить маршрут до эндпоинта
func (h *userHandler) Setup(router huma.API) {
	huma.Register(router, huma.Operation{
		OperationID: "login",
		Path:        "/api/login",
		Method:      http.MethodPost,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"auth",
		},
		Summary:     "Login",
		Description: "Получить токен для взаимодействия. Нуждается в InitDataRaw строке из Telegram Mini App. Действует 1 час",
	}, h.login)
}
