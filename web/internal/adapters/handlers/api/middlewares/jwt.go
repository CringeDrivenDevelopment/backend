package middlewares

import (
	"backend/cmd/app"
	"backend/internal/domain/service"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
	"time"
)

type MiddlewareHandler struct {
	userService  *service.UserService
	tokenService *service.TokenService
	api          huma.API
	logger       *zap.Logger
}

const UserJwtKey = "user"

// NewMiddlewareHandler is a function that returns a new instance of MiddlewareHandler.
func NewMiddlewareHandler(app *app.App) *MiddlewareHandler {
	userService := service.NewUserService(app.DB)
	tokenService := service.NewTokenService(app.Settings.JwtSecret, time.Hour)

	return &MiddlewareHandler{
		userService:  userService,
		tokenService: tokenService,
		api:          app.Router,
		logger:       app.Logger,
	}
}

// IsAuthenticated is a function that checks whether the user has sufficient rights to access the endpoint
/*
 * tokenType string - the type of token that is required to access the endpoint
 * requiredRights ...string - the rights that the user must have
 */
func (h *MiddlewareHandler) IsAuthenticated(ctx huma.Context, next func(ctx huma.Context)) {
	authHeader := ctx.Header("Authorization")

	// user, err := h.tokenService.GetUserFromJWT(authHeader, ctx.Context(), h.userService.GetByID)
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

	// Otherwise, just continue as normal.
	next(ctx)
}
