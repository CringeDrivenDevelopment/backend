package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"slices"
)

type PlaylistService struct {
	pool *pgxpool.Pool
}

func NewPlaylistService(app *app.App) *PlaylistService {
	return &PlaylistService{pool: app.DB}
}

func (s *PlaylistService) Create(ctx context.Context, title string, source string) (dto.Playlist, error) {
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
		Type:          source,
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
		AllowedIds:    nil,
		AllowedLength: 0,
		Role:          dto.OwnerRole,
		Type:          source,
	}, nil
}

func (s *PlaylistService) GetById(ctx context.Context, playlistId string, userId int64) (dto.Playlist, error) {
	queries := repository.New(s.pool)
	playlist, err := queries.GetPlaylistById(ctx, repository.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return dto.Playlist{}, err
	}

	if !slices.Contains(dto.UserRoles, playlist.Role) {
		return dto.Playlist{}, errors.New("playlist not found")
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

	count := playlist.Count.Int32
	allowedCount := playlist.AllowedCount.Int32
	time := playlist.Time
	allowedTime := playlist.AllowedTime

	return dto.Playlist{
		Id:            playlist.ID,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        tracks,
		AllowedIds:    playlist.AllowedTracks,
		Count:         int(count),
		Length:        int(time),
		AllowedCount:  int(allowedCount),
		AllowedLength: int(allowedTime),
		Role:          playlist.Role,
		Type:          playlist.Type,
	}, nil
}

func (s *PlaylistService) GetAll(ctx context.Context, userId int64) ([]dto.Playlist, error) {
	queries := repository.New(s.pool)
	playlists, err := queries.GetUserPlaylists(ctx, userId)
	if err != nil {
		return nil, err
	}

	result := make([]dto.Playlist, len(playlists))
	for i, playlist := range playlists {
		count := playlist.Count.Int32
		allowedCount := playlist.AllowedCount.Int32
		time := playlist.Time
		allowedTime := playlist.AllowedTime

		result[i] = dto.Playlist{
			Id:            playlist.ID,
			Title:         playlist.Title,
			Thumbnail:     playlist.Thumbnail,
			Count:         int(count),
			Length:        int(time),
			AllowedCount:  int(allowedCount),
			AllowedLength: int(allowedTime),
			Tracks:        make([]dto.Track, 0),
			AllowedIds:    make([]string, 0),
			Role:          playlist.Role,
			Type:          playlist.Type,
		}
	}

	return result, nil
}

func (s *PlaylistService) Rename(ctx context.Context, playlistId string, userId int64, title string) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	playlist, err := queries.GetPlaylistById(ctx, repository.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	playlist.Title = title

	err = queries.EditPlaylist(ctx, repository.EditPlaylistParams{
		ID:            playlist.ID,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        playlist.Tracks,
		AllowedTracks: playlist.AllowedTracks,
		Type:          playlist.Type,
	})
	if err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}

func (s *PlaylistService) Delete(ctx context.Context, playlistId string) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	err := queries.DeletePlaylist(ctx, playlistId)
	if err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
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
