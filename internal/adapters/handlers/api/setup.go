package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/middlewares"
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(app *app.App) {
	app.Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://tg-mini-app.local",
			"https://cloud.lxft.tech",
			"https://lxft.tech",
			"http://localhost",
		},
	}))

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
	}, func(ctx context.Context, input *struct{}) (*struct {
		Body struct {
			Status string `json:"status" example:"Pong!"`
		}
	}, error) {
		resp := &struct{ Body struct{ Status string } }{}
		resp.Body.Status = "Pong!"
		return (*struct {
			Body struct {
				Status string `json:"status" example:"Pong!"`
			}
		})(resp), nil
	})

	middlewareHandler := middlewares.NewMiddlewareHandler(app)
	//
	// Setup user routes
	newUserHandler(app).Setup(app.Router)

	newPlaylistsHandler(app).Setup(app.Router, middlewareHandler.IsAuthenticated)

	newTrackHandler(app).Setup(app.Router, middlewareHandler.IsAuthenticated)
}
