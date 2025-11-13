package main

import (
	"backend/internal/infra"
	"backend/internal/service"
	"backend/internal/transport/api/handlers"
	"backend/internal/transport/api/middlewares"
	"backend/pkg/youtube"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

func main() {
	// TODO: log db requests
	// TODO: add otel
	// TODO: add image proxy, DL

	fx.New(
		fx.Provide(
			infra.NewLogger,
			infra.NewConfig,
			infra.NewPostgresConnection,
			infra.NewEcho,
			infra.NewHuma,
			youtube.New,
			service.NewAuth,
			service.NewPermission,
			service.NewPlaylist,
			service.NewTrack,
			service.NewUser,
			middlewares.NewLogger,
			middlewares.NewAuth,
			handlers.NewAuth,
			handlers.NewPlaylist,
			handlers.NewTrack,
		),
		fx.Invoke(func(echo *echo.Echo) {}),
	).Run()
}
