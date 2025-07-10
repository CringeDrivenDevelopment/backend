package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/middlewares"
	"backend/internal/domain/dto"
	"backend/internal/domain/service"
	"context"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type trackHandler struct {
	playlistService *service.PlaylistService
	trackService    *service.TrackService
	youtubeService  *service.YoutubeService
}

func newTrackHandler(app *app.App) *trackHandler {
	return &trackHandler{
		playlistService: service.NewPlaylistService(app),
		trackService:    service.NewTrackService(app),
		youtubeService:  service.NewYoutubeService(app),
	}
}

func (h *trackHandler) submit(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	err := h.trackService.Submit(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	err = h.youtubeService.Download(context.Background(), input.TrackId)
	if err != nil {
		return nil, huma.Error500InternalServerError("Youtube download failed", err)
	}

	return nil, nil
}

func (h *trackHandler) remove(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	playlist, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if playlist.Role != dto.ModeratorRole && playlist.Role != dto.OwnerRole {
		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.RemoveApproved(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to remove from allowed tracks", err)
	}

	err = h.trackService.Decline(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to remove from tracks", err)
	}

	return nil, nil
}

func (h *trackHandler) approve(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	playlist, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if playlist.Role != dto.ModeratorRole && playlist.Role != dto.OwnerRole {
		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.Approve(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return nil, nil
}

func (h *trackHandler) decline(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}

	playlist, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if playlist.Role != dto.ModeratorRole && playlist.Role != dto.OwnerRole {
		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.Decline(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return nil, nil
}

func (h *trackHandler) Setup(router huma.API, auth func(ctx huma.Context, next func(ctx huma.Context))) {
	huma.Register(router, huma.Operation{
		OperationID: "tracks-submit",
		Path:        "/api/playlists/{playlist_id}/{track_id}/submit",
		Method:      http.MethodPost,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Add a track to playlist",
		Description: "If your role is viewer, you are putting track for review. If you're owner/moderator you're automatically adding track to allowed",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.submit)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-remove",
		Path:        "/api/playlists/{playlist_id}/{track_id}/remove",
		Method:      http.MethodDelete,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Remove allowed track from playlist",
		Description: "If your role is viewer, you are not allowed. If you're owner/moderator you can invoke this",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.remove)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-approve",
		Path:        "/api/playlists/{playlist_id}/{track_id}/approve",
		Method:      http.MethodPatch,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Approve track in playlist",
		Description: "If your role is viewer, you are not allowed. If you're owner/moderator you can invoke this",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.approve)

	huma.Register(router, huma.Operation{
		OperationID: "tracks-decline",
		Path:        "/api/playlists/{playlist_id}/{track_id}/decline",
		Method:      http.MethodDelete,
		Errors: []int{
			401,
			404,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Decline a track in submissions",
		Description: "If your role is viewer, you are not allowed. If you're owner/moderator you can invoke this",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.decline)
}
