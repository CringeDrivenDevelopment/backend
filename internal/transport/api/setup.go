package api

import (
	"backend/internal/app"
	handlers2 "backend/internal/transport/api/handlers"
	middlewares2 "backend/internal/transport/api/middlewares"
	"net/http"

	"github.com/labstack/echo/v4"
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
	// TODO: log db requests
	// TODO: add otel
	app.Server.Use(middlewares2.ZapLogger(app.Logger))

	// recover from panic
	if !app.Settings.Debug {
		app.Server.Use(middleware.Recover())
	}

	// Provide a minimal config for startup check
	app.Server.GET("/api/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	authMiddleware := middlewares2.NewAuth(app)

	// TODO: add image proxy, sound proxy, DL

	// Добавить маршруты до основного API
	handlers2.NewAuth(app).Setup(app.Api)
	handlers2.NewPlaylist(app).Setup(app.Api, authMiddleware.IsAuthenticated)
	handlers2.NewTrack(app).Setup(app.Api, authMiddleware.IsAuthenticated)
}
