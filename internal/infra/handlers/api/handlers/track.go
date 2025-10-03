package handlers

import (
	"backend/internal/app"
	"backend/internal/app/service"
	"backend/internal/infra/database/queries"
	"backend/internal/infra/handlers/api/dto"
	"backend/internal/infra/handlers/api/middlewares"
	"context"
	"errors"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Track struct {
	playlistService *service.Playlist
	trackService    *service.Track

	logger *zap.Logger
}

func NewTrack(app *app.App) *Track {
	return &Track{
		playlistService: service.NewPlaylistService(app),
		trackService:    service.NewTrackService(app),
		logger:          app.Logger,
	}
}

// search - поиск трека по query, требуется длина более
func (h *Track) search(ctx context.Context, input *struct {
	Query string `query:"query"`
}) (*struct {
	Body []dto.Track
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}
	query := input.Query

	h.logger.Warn(fmt.Sprintf("track search: user_id - %d, query - %s", val, query))

	if len(input.Query) < 3 {
		return nil, huma.Error400BadRequest("query too short")
	}

	search, err := h.trackService.Search(ctx, query, val)
	if err != nil {
		h.logger.Error("search error: "+query, zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &struct{ Body []dto.Track }{Body: search}, nil
}

func (h *Track) submit(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ" doc:"track id"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}

	h.logger.Info(fmt.Sprintf("playlist submit: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

	_, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist not found: user_id - %d, playlist_id - %s", val, input.PlaylistId))

			return nil, huma.Error404NotFound("playlist not found", err)
		}

		h.logger.Error(fmt.Sprintf("playlist submit error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	err = h.trackService.Submit(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("track not found", err)
		}

		h.logger.Error(fmt.Sprintf("playlist submit error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return nil, nil
}

func (h *Track) unapprove(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ" doc:"track id"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}

	entity, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist not found: user_id - %d, playlist_id - %s", val, input.PlaylistId))

			return nil, huma.Error404NotFound("playlist not found", err)
		}

		h.logger.Error(fmt.Sprintf("unapprove error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	if entity.Role != queries.PlaylistRoleOwner && entity.Role != queries.PlaylistRoleModerator {
		h.logger.Warn(fmt.Sprintf("playlist unapprove not allowed: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.Unapprove(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist track not found: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

			return nil, huma.Error404NotFound("track not found", err)
		}

		h.logger.Error(fmt.Sprintf("unapprove error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return nil, nil
}

func (h *Track) approve(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ" doc:"track id"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}

	h.logger.Info(fmt.Sprintf("playlist approve: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

	entity, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist not found: user_id - %d, playlist_id - %s", val, input.PlaylistId))

			return nil, huma.Error404NotFound("playlist not found", err)
		}

		h.logger.Error(fmt.Sprintf("approve error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	if entity.Role != queries.PlaylistRoleOwner && entity.Role != queries.PlaylistRoleModerator {
		h.logger.Warn(fmt.Sprintf("playlist approve not allowed: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.Approve(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist track not found: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

			return nil, huma.Error404NotFound("track not found", err)
		}

		h.logger.Error(fmt.Sprintf("approve error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return nil, nil
}

func (h *Track) decline(ctx context.Context, input *struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ" doc:"track id"`
}) (*struct{}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}

	h.logger.Info(fmt.Sprintf("playlist decline: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

	entity, err := h.playlistService.GetById(ctx, input.PlaylistId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist not found: user_id - %d, playlist_id - %s", val, input.PlaylistId))

			return nil, huma.Error404NotFound("playlist not found", err)
		}

		h.logger.Error(fmt.Sprintf("decline error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	if entity.Role != queries.PlaylistRoleOwner && entity.Role != queries.PlaylistRoleModerator {
		h.logger.Warn(fmt.Sprintf("playlist decline not allowed: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

		return nil, huma.Error403Forbidden("action not allowed")
	}

	err = h.trackService.Decline(ctx, input.PlaylistId, input.TrackId, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist track not found: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId))

			return nil, huma.Error404NotFound("track not found", err)
		}

		h.logger.Error(fmt.Sprintf("decline error: user_id - %d, playlist_id - %s, track_id - %s", val, input.PlaylistId, input.TrackId), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return nil, nil
}
