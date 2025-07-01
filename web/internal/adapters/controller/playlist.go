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
	service *service.PlaylistService
}

func newPlaylistsHandler(app *app.App) *playlistHandler {
	return &playlistHandler{service: service.NewPlaylistService(app)}
}

func (h *playlistHandler) create(ctx context.Context, input *struct {
	Body struct {
		Title string `json:"title"`
	}
}) (*struct {
	Body dto.Playlist
}, error) {
	resp, err := h.service.Create(ctx, input.Body.Title)
	if err != nil {
		return nil, err
	}
	return &struct{ Body dto.Playlist }{Body: resp}, nil
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
	}, h.create)
}
