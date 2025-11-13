package main

import (
	"backend/internal/infra"
	"backend/internal/service"
	"backend/internal/transport/api/handlers"
	"backend/internal/transport/api/middlewares"
	"backend/pkg/youtube"

	"go.uber.org/fx"
)

func main() {
	// TODO: log db requests
	// TODO: add otel
	// TODO: add image proxy, DL

	fx.New(
		fx.Provide(
			// REST API
			infra.NewEcho,
			infra.NewHuma,
			middlewares.NewLogger,
			middlewares.NewAuth,
			handlers.NewAuth,
			handlers.NewPlaylist,
			handlers.NewTrack,

			// services and infra
			infra.NewLogger,
			infra.NewConfig,
			infra.NewPostgresConnection,
			youtube.New,
			service.NewAuth,
			service.NewPermission,
			service.NewPlaylist,
			service.NewTrack,
			service.NewUser,
		),
		fx.Invoke(func(auth *handlers.Auth, track *handlers.Track, playlist *handlers.Playlist) {
			// need echo and huma to start the api

			// need each of controllers, to register them, maybe i'll use hooks

			// no need to call infra, apis and services, they're deps, started automatically
		}),
	).Run()
}
