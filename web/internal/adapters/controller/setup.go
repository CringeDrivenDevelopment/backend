package controller

import (
	"backend/cmd/app"
	"backend/internal/domain/dto"
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

func Setup(app *app.App) {
	app.Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))
	// TODO: change origin to strict mode

	// log all requests
	// TODO: to zap logger
	// TODO: log db requests
	// TODO: add otel
	app.Server.Use(middleware.Logger())

	// recover from panic
	app.Server.Use(middleware.Recover())

	// Provide a minimal config for startup check
	huma.Register(app.Router, huma.Operation{
		OperationID: "ping",
		Method:      http.MethodGet,
		Path:        "/api/ping",
		Summary:     "Pong!",
		Description: "Check if server has started and running",
		Tags:        []string{"ping"},
	}, func(ctx context.Context, input *struct{}) (*dto.PingOutput, error) {
		resp := &dto.PingOutput{}
		resp.Body.Status = "Pong!"
		return resp, nil
	})

	// middlewareHandler := middlewares.NewMiddlewareHandler(app)
	//
	// Setup user routes
	newUserHandler(app).Setup(app.Router)

	newTracksHandler(app).Setup(app.Router)
	// TODO: add middleware

	newPlaylistsHandler(app).Setup(app.Router)
}
