package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/middlewares"
	"backend/internal/domain/notification"
	"backend/internal/domain/permission"
	"backend/internal/domain/playlist"
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

// TODO: edit playlist title, check for changes if playlist is not custom

type playlistHandler struct {
	playlistService     *playlist.Service
	permissionService   *permission.Service
	notificationService *notification.Service

	logger *zap.Logger
}

func newPlaylistsHandler(app *app.App) *playlistHandler {
	return &playlistHandler{
		playlistService:     playlist.NewService(app),
		permissionService:   permission.NewService(app),
		notificationService: notification.NewService(app),
		logger:              app.Logger,
	}
}

func (h *playlistHandler) getById(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct {
	Body playlist.DtoPlaylist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.GetById(ctx, input.Id, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}
	return &struct{ Body playlist.DtoPlaylist }{Body: resp}, nil
}

func (h *playlistHandler) getAll(ctx context.Context, _ *struct{}) (*struct {
	Body []playlist.DtoPlaylist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.GetAll(ctx, val)
	if err != nil {
		return nil, huma.Error500InternalServerError("Playlists not found", err)
	}

	return &struct{ Body []playlist.DtoPlaylist }{Body: resp}, nil
}

func (h *playlistHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "playlist-by-id",
		Path:        "/api/playlists/{id}",
		Method:      http.MethodGet,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Get playlist by id",
		Description: "TODO: Change",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.getById)

	huma.Register(router, huma.Operation{
		OperationID: "playlist-all",
		Path:        "/api/playlists",
		Method:      http.MethodGet,
		Errors: []int{
			401,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Get all",
		Description: "Get all playlists of user",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.getAll)
}
