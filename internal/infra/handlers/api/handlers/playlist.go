package handlers

import (
	"backend/internal/app"
	"backend/internal/app/service"
	"backend/internal/infra/handlers/api/dto"
	"backend/internal/infra/handlers/api/middlewares"
	"context"
	"errors"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Playlist struct {
	playlistService     *service.Playlist
	permissionService   *service.Permission
	notificationService *service.TgNotification

	logger *zap.Logger
}

// NewPlaylist - создать новый экземпляр обработчика
func NewPlaylist(app *app.App) *Playlist {
	return &Playlist{
		playlistService:     service.NewPlaylistService(app),
		permissionService:   service.NewPermissionService(app),
		notificationService: service.NewNotificationService(app),
		logger:              app.Logger,
	}
}

// getById - получить плейлист по ULID
func (h *Playlist) getById(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*struct {
	Body dto.Playlist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}

	resp, err := h.playlistService.GetById(ctx, input.Id, val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.logger.Info(fmt.Sprintf("playlist not found: user_id - %d, playlist_id - %s", val, input.Id))

			return nil, huma.Error404NotFound("playlist not found", err)
		}

		h.logger.Error(fmt.Sprintf("get playlist error: user_id - %d, playlist_id - %s", val, input.Id), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &struct{ Body dto.Playlist }{Body: resp}, nil
}

// getAll - получить список плейлистов для пользователя
func (h *Playlist) getAll(ctx context.Context, _ *struct{}) (*struct {
	Body []dto.Playlist
}, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, huma.Error500InternalServerError("internal server error")
	}

	resp, err := h.playlistService.GetAll(ctx, val)
	if err != nil {
		h.logger.Error(fmt.Sprintf("get all playlists error: user_id - %d", val), zap.Error(err))

		return nil, huma.Error500InternalServerError("internal server error")
	}

	return &struct{ Body []dto.Playlist }{Body: resp}, nil
}
