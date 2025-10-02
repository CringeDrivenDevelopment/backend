package rest

import (
	"backend/internal/application"
	"backend/internal/interfaces/rest/middlewares"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Setup(app *application.App) {
	app.Server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://tg-mini-app.local",
			"https://cloud.lxft.tech",
			"https://lxft.tech",
			"http://localhost",
		},
	}))

	// log all requests
	// TODO: log db requests
	// TODO: add otel
	app.Server.Use(middlewares.ZapLogger(app.Logger))

	// recover from panic
	app.Server.Use(middleware.Recover())

	// Provide a minimal config for startup check
	app.Server.GET("/api/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	middlewareHandler := middlewares.NewAuth(app)
	//
	// Setup user routes
	newUserHandler(app).Setup(app.Api)

	newPlaylistsHandler(app).Setup(app.Api, middlewareHandler.IsAuthenticated)

	newTrackHandler(app).Setup(app.Api, middlewareHandler.IsAuthenticated)
}
