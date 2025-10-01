package api

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api/middlewares"
	"backend/internal/domain/playlist"
	"backend/internal/domain/track"
	"backend/internal/infra/database/queries"
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type trackHandler struct {
	playlistService *playlist.Service
	trackService    *track.Service
}

func newTrackHandler(app *app.App) *trackHandler {
	return &trackHandler{
		playlistService: playlist.NewService(app),
		trackService:    track.NewService(app),
	}
}

func (h *trackHandler) search(ctx context.Context, input *struct {
	Query string `query:"query"`
}) (*struct {
	Body []track.DtoTrack
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		return nil, huma.Error500InternalServerError("User not found in context")
	}
	query := input.Query

	if len(input.Query) < 3 {
		return nil, huma.Error400BadRequest("query too short")
	}

	search, err := h.trackService.Search(ctx, query, val)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error(), err)
	}

	return &struct{ Body []track.DtoTrack }{Body: search}, nil
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

	entity, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if entity.Role != queries.PlaylistRoleOwner && entity.Role != queries.PlaylistRoleModerator {
		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.Unapprove(ctx, input.PlaylistId, input.TrackId, val)
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

	entity, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if entity.Role != queries.PlaylistRoleOwner && entity.Role != queries.PlaylistRoleModerator {
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

	entity, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		return nil, huma.Error404NotFound("playlist not found", err)
	}

	if entity.Role != queries.PlaylistRoleOwner && entity.Role != queries.PlaylistRoleModerator {
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
		OperationID: "track-search",
		Path:        "/api/track/search",
		Method:      http.MethodGet,
		Errors: []int{
			400,
			401,
			500,
		},
		Tags: []string{
			"tracks",
		},
		Summary:     "Search",
		Description: "Find tracks by query",
		Middlewares: huma.Middlewares{auth},
		Security: []map[string][]string{
			{
				"jwt": []string{},
			},
		},
	}, h.search)

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
