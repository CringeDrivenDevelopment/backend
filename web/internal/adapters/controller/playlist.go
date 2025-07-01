package controller

import (
	"backend/cmd/app"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type playlistHandler struct {
	playlistService *service.PlaylistService
}

func newPlaylistsHandler(app *app.App) *playlistHandler {
	return &playlistHandler{playlistService: service.NewPlaylistService(app)}
}

func (h *playlistHandler) create(ctx context.Context, input *struct {
	Body struct {
		Title string `json:"title"`
	}
}) (*struct {
	Body dto.Playlist
}, error) {
	resp, err := h.playlistService.Create(ctx, input.Body.Title)
	if err != nil {
		return nil, err
	}
	return &struct{ Body dto.Playlist }{Body: resp}, nil
}

func (h *playlistHandler) getById(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct {
	Body dto.Playlist
}, error) {
	resp, err := h.playlistService.GetById(ctx, input.Id)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}
	return &struct{ Body dto.Playlist }{Body: resp}, nil
}

func (h *playlistHandler) submit(ctx context.Context, input *struct {
	Body struct {
		TrackId string `json:"track_id"`
	}
}) (*struct {
	Body dto.Playlist
}, error) {

	return nil, nil
}

func (h *playlistHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "create-playlist",
		Path:        "/api/playlist/new",
		Method:      http.MethodPost,
		Errors: []int{
			400,
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
		OperationID: "get-playlist-by-id",
		Path:        "/api/playlist/{id}",
		Method:      http.MethodGet,
		Errors: []int{
			404,
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
}
