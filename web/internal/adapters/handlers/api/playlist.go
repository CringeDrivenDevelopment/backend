package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/middlewares"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
	"net/http"
)

// TODO: edit playlist title, check for changes if playlist is not custom

type playlistHandler struct {
	playlistService     *service.PlaylistService
	permissionService   *service.PermissionService
	youtubeService      *service.YoutubeService
	notificationService *service.TelegramNotificationService

	logger *zap.Logger
}

func newPlaylistsHandler(app *app.App) *playlistHandler {
	return &playlistHandler{
		playlistService:     service.NewPlaylistService(app),
		permissionService:   service.NewPermissionService(app),
		youtubeService:      service.NewYoutubeService(app),
		notificationService: service.NewTelegramNotificationService(app),
		logger:              app.Logger,
	}
}

func (h *playlistHandler) create(ctx context.Context, input *struct {
	Body struct {
		Title string `json:"title"`
	}
}) (*struct {
	Body dto.Playlist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.Create(ctx, input.Body.Title, dto.CustomSource)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create playlist", err)
	}
	err = h.permissionService.Add(ctx, dto.OwnerRole, resp.Id, val)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to add playlist to user", err)
	}

	return &struct{ Body dto.Playlist }{Body: resp}, nil
}

func (h *playlistHandler) getById(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct {
	Body dto.Playlist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.GetById(ctx, input.Id, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}
	return &struct{ Body dto.Playlist }{Body: resp}, nil
}

func (h *playlistHandler) download(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct {
	Body dto.Archive
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.GetById(ctx, input.Id, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	archive, err := h.youtubeService.Archive(ctx, resp.AllowedIds)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to download playlist", err)
	}

	// TODO: FIX THIS HARDCODED HELL
	err = h.notificationService.Send(val, "Привет! Вот ссылка на твой архив с песнями, если через миниапп не работает. Удачно потусить)"+"\n\n"+fmt.Sprintf("https://cloud.lxft.tech/api/youtube/%s", archive.Filename))
	if err != nil {
		h.logger.Warn(err.Error())
	}

	return &struct{ Body dto.Archive }{Body: archive}, nil
}

func (h *playlistHandler) delete(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	playlist, err := h.playlistService.GetById(ctx, input.Id, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if playlist.Role != dto.OwnerRole || playlist.Type != dto.CustomSource {
		return nil, huma.Error403Forbidden("you're not allowed to do this", err)
	}

	err = h.playlistService.Delete(ctx, input.Id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *playlistHandler) getAll(ctx context.Context, _ *struct{}) (*struct {
	Body []dto.Playlist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.GetAll(ctx, val)
	if err != nil {
		return nil, huma.Error500InternalServerError("Playlists not found", err)
	}

	return &struct{ Body []dto.Playlist }{Body: resp}, nil
}

func (h *playlistHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "playlist-create",
		Path:        "/api/playlists/new",
		Method:      http.MethodPost,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Create new playlist",
		Description: "TODO: Change",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.create)

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
		OperationID: "playlist-download",
		Path:        "/api/playlists/{id}/download",
		Method:      http.MethodPost,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Download",
		Description: "Get playlist archive.zip url",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.download)

	huma.Register(router, huma.Operation{
		OperationID: "playlist-delete",
		Path:        "/api/playlists/{id}",
		Method:      http.MethodDelete,
		Errors: []int{
			401,
			403,
			404,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Delete a playlist",
		Description: "TODO: Change",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.delete)

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
