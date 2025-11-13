package infra

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewEcho(lc fx.Lifecycle, cfg *Config, logger *zap.Logger, loggerWare echo.MiddlewareFunc) *echo.Echo {
	router := echo.New()

	if !cfg.Debug {
		router.Use(middleware.Recover())
	}

	router.GET("/api/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	router.HideBanner = true
	router.HidePort = true

	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://tg-mini-app.local",
			"https://cloud.lxft.tech",
			"https://lxft.tech",
			"http://localhost",
		},
	}))

	router.Use(loggerWare)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting server on :8080")
			go func() {
				err := router.Start(":8080")
				if err != nil {
					logger.Fatal("stopping server, cause: error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopped server")
			return router.Shutdown(ctx)
		},
	})

	return router
}
