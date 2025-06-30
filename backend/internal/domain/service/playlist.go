package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/oklog/ulid/v2"
)

type PlaylistService struct {
	pool       *pgxpool.Pool
	s3         *minio.Client
	bucketName string
}

func NewPlaylistService(app *app.App) *PlaylistService {
	return &PlaylistService{pool: app.DB}
}

func (s *PlaylistService) Create(ctx context.Context, title string) (dto.Playlist, error) {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return dto.Playlist{}, txErr
	}

	id := ulid.Make().String()

	queries := repository.New(tx)
	if err := queries.CreatePlaylist(ctx, repository.CreatePlaylistParams{
		ID:            id,
		Title:         title,
		Thumbnail:     "",
		Tracks:        make([]string, 0),
		AllowedTracks: make([]string, 0),
	}); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return dto.Playlist{}, txErr
		}
		return dto.Playlist{}, err
	}

	return dto.Playlist{
		Id:            id,
		Title:         title,
		Thumbnail:     "",
		Tracks:        nil,
		Length:        0,
		AllowedTracks: nil,
		AllowedLength: 0,
	}, nil
}
