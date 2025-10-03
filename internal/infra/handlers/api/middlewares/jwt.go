package middlewares

import (
	"backend/internal/application"
	service2 "backend/internal/domain/service"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

type Auth struct {
	userService  *service2.User
	tokenService *service2.Auth
	api          huma.API
	logger       *zap.Logger
}

const UserJwtKey = "user"

// NewAuth - создать новый обработчик для middleware
func NewAuth(app *application.App) *Auth {
	userService := service2.NewUserService(app)
	tokenService := service2.NewAuthService(app, time.Hour)

	return &Auth{
		userService:  userService,
		tokenService: tokenService,
		api:          app.Api,
		logger:       app.Logger,
	}
}

// IsAuthenticated - проверить, авторизован ли пользователь для выполнения запроса
func (h *Auth) IsAuthenticated(ctx huma.Context, next func(ctx huma.Context)) {
	authHeader := ctx.Header("Authorization")

	// проверить токен
	id, err := h.tokenService.VerifyToken(authHeader)
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
