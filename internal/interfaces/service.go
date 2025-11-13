package interfaces

import (
	"backend/internal/infra/queries"
	"backend/internal/transport/api/dto"
	"backend/internal/transport/bot/models"
	"context"
)

type UserService interface {
	Create(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) error
}

type AuthService interface {
	VerifyToken(authHeader string) (int64, error)
	GenerateToken(userID int64) (string, error)
	ParseInitData(initDataRaw string) (int64, error)
}

type PlaylistService interface {
	Create(ctx context.Context, title string, playlistType queries.PlaylistType, telegramId int64) (dto.Playlist, error)
	GetByGroup(ctx context.Context, telegramId int64) (dto.Playlist, error)
	GetById(ctx context.Context, playlistId string, userId int64) (dto.Playlist, error)
	GetAll(ctx context.Context, userId int64) ([]dto.Playlist, error)
	Rename(ctx context.Context, playlistId string, title string, userId int64) error
	UpdatePhoto(ctx context.Context, playlistId string, thumbnail string, userId int64) error
	Delete(ctx context.Context, playlistId string) error
}

type PermissionService interface {
	Add(ctx context.Context, role queries.PlaylistRole, playlist string, userId int64) error
	AddGroup(ctx context.Context, playlist string, users []models.ParticipantData) error
	Remove(ctx context.Context, playlist string, userId int64) error
	Edit(ctx context.Context, role queries.PlaylistRole, playlist string, userId int64) error
	Get(ctx context.Context, userId int64, role queries.PlaylistRole) (string, error)
}

type TrackService interface {
	Search(ctx context.Context, query string) ([]dto.Track, error)
	GetById(ctx context.Context, id string) (dto.Track, error)
	Approve(ctx context.Context, playlistId string, trackId string, userId int64) error
	Decline(ctx context.Context, playlistId string, trackId string, userId int64) error
	Submit(ctx context.Context, playlistId string, trackId string, userId int64) error
	Unapprove(ctx context.Context, playlistId string, trackId string, userId int64) error
}
