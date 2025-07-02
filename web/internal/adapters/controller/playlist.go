package controller

import (
	"backend/cmd/app"
	"backend/internal/adapters/controller/middlewares"
	"backend/internal/adapters/repository"
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
	val, ok := ctx.Value(middlewares.USER_JWT_KEY).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.Create(ctx, input.Body.Title, val.ID)
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
	val, ok := ctx.Value(middlewares.USER_JWT_KEY).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	resp, err := h.playlistService.GetById(ctx, input.Id, val.ID)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}
	return &struct{ Body dto.Playlist }{Body: resp}, nil
}

func (h *playlistHandler) submit(ctx context.Context, input *struct {
	PlaylistId string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	Body       struct {
		TrackId string `json:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ"`
	}
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.USER_JWT_KEY).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	err := h.playlistService.SubmitTrack(ctx, input.PlaylistId, input.Body.TrackId, val.ID)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}
	return nil, nil
}

func (h *playlistHandler) delete(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.USER_JWT_KEY).(repository.User)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	_, err := h.playlistService.GetById(ctx, input.Id, val.ID)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	err = h.playlistService.Delete(ctx, input.Id)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *playlistHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "playlist-create",
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
		OperationID: "playlist-submit",
		Path:        "/api/playlist/{id}/submit",
		Method:      http.MethodPost,
		Errors: []int{
			404,
			500,
		},
		Tags: []string{
			"playlist",
		},
		Summary:     "Add a track to playlist for a review (if role==viewer)",
		Description: "TODO: Change",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.submit)

	huma.Register(router, huma.Operation{
		OperationID: "playlist-delete",
		Path:        "/api/playlist/{id}",
		Method:      http.MethodDelete,
		Errors: []int{
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
}
