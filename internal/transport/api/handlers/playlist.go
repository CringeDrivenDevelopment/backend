package handlers

import (
	"backend/internal/app"
	"backend/internal/errorz"
	"backend/internal/service"
	"backend/internal/transport/api/dto"
	"backend/internal/transport/api/middlewares"
	"context"
	"errors"
	"fmt"

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

// getById - получить плейлист по ID
func (h *Playlist) getById(ctx context.Context, input *struct {
	Id string `path:"id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
}) (*dto.PlaylistByIdResponse, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, errorz.Convert(errors.New("token not found in context"))
	}

	h.logger.Info(fmt.Sprintf("playlistById: user_id - %d, playlist_id - %s", val, input.Id))

	resp, err := h.playlistService.GetById(ctx, input.Id, val)
	if err != nil {
		h.logger.Error(fmt.Sprintf("playlistById error: user_id - %d, playlist_id - %s", val, input.Id), zap.Error(err))

		return nil, errorz.Convert(err)
	}

	return &dto.PlaylistByIdResponse{Body: resp}, nil
}

// getAll - получить список плейлистов для пользователя
func (h *Playlist) getAll(ctx context.Context, _ *struct{}) (*dto.PlaylistsResponse, error) {
	val, ok := ctx.Value(middlewares.UserJwtKey).(int64)
	if !ok {
		h.logger.Error("user not found in context")

		return nil, errorz.Convert(errors.New("token not found in context"))
	}

	h.logger.Info(fmt.Sprintf("playlists: user_id - %d", val))

	resp, err := h.playlistService.GetAll(ctx, val)
	if err != nil {
		h.logger.Error(fmt.Sprintf("playlists error: user_id - %d", val), zap.Error(err))

		return nil, errorz.Convert(err)
	}

	return &dto.PlaylistsResponse{Body: resp}, nil
}
