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

	if err := tx.Commit(ctx); err != nil {
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

func (s *PlaylistService) GetById(ctx context.Context, id string) (dto.Playlist, error) {
	queries := repository.New(s.pool)
	playlist, err := queries.GetPlaylistById(ctx, id)
	if err != nil {
		return dto.Playlist{}, err
	}

	tracks := make([]dto.Track, len(playlist.Tracks))
	for i, track := range playlist.Tracks {
		dbTrack, err := queries.GetTrackById(ctx, track)
		if err != nil {
			return dto.Playlist{}, err
		}

		tracks[i] = dto.Track{
			Id:        dbTrack.ID,
			Title:     dbTrack.Title,
			Authors:   dbTrack.Authors,
			Explicit:  dbTrack.Explicit,
			Length:    dbTrack.Length,
			Thumbnail: dbTrack.Thumbnail,
		}
	}

	allowedTracks := make([]dto.Track, len(playlist.AllowedTracks))
	for i, track := range playlist.AllowedTracks {
		dbTrack, err := queries.GetTrackById(ctx, track)
		if err != nil {
			return dto.Playlist{}, err
		}

		allowedTracks[i] = dto.Track{
			Id:        dbTrack.ID,
			Title:     dbTrack.Title,
			Authors:   dbTrack.Authors,
			Explicit:  dbTrack.Explicit,
			Length:    dbTrack.Length,
			Thumbnail: dbTrack.Thumbnail,
		}
	}

	return dto.Playlist{
		Id:            playlist.ID,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Length:        len(playlist.Tracks),
		Tracks:        tracks,
		AllowedTracks: allowedTracks,
		AllowedLength: len(playlist.AllowedTracks),
	}, nil
}

func (s *PlaylistService) SubmitTrack(ctx context.Context, playlistId, trackId string) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	playlist, err := queries.GetPlaylistById(ctx, playlistId)
	if err != nil {
		return err
	}

	if _, err := queries.GetTrackById(ctx, trackId); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
		return err
	}

	err = queries.EditPlaylist(ctx, repository.EditPlaylistParams{
		ID:            playlistId,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        append(playlist.Tracks, trackId),
		AllowedTracks: playlist.AllowedTracks,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}
