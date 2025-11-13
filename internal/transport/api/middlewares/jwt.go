package middlewares

import (
	"backend/internal/app"
	"backend/internal/service"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

type Auth struct {
	userService *service.User
	authService *service.Auth
	api         huma.API
	logger      *zap.Logger
}

const UserJwtKey = "user"

// NewAuth - создать новый обработчик для middleware
func NewAuth(app *app.App) *Auth {
	userService := service.NewUserService(app)
	authService := service.NewAuthService(app, time.Hour)

	return &Auth{
		userService: userService,
		authService: authService,
		api:         app.Api,
		logger:      app.Logger,
	}
}

// IsAuthenticated - проверить, авторизован ли пользователь для выполнения запроса
func (h *Auth) IsAuthenticated(ctx huma.Context, next func(ctx huma.Context)) {
	authHeader := ctx.Header("Authorization")

	// проверить токен
	id, err := h.authService.VerifyToken(authHeader)
	if err != nil {
		err := huma.WriteErr(h.api, ctx, 401, "unauthorized")
		if err != nil {
			h.logger.Error("failed to return status 401 from middleware: " + err.Error())
			return
		}
		return
	}

	ctx = huma.WithValue(ctx, UserJwtKey, id)

	// продолжить выполнение запроса
	next(ctx)
}
