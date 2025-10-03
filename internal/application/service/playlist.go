package service

import (
	"backend/internal/application"
	"backend/internal/infra/database/queries"
	dto2 "backend/internal/infra/handlers/api/dto"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

type Playlist struct {
	pool *pgxpool.Pool
}

func NewPlaylistService(app *application.App) *Playlist {
	return &Playlist{pool: app.DB}
}

func (s *Playlist) Create(ctx context.Context, title string, playlistType queries.PlaylistType) (dto2.Playlist, error) {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return dto2.Playlist{}, txErr
	}

	id := ulid.Make().String()

	q := queries.New(tx)

	if err := q.CreatePlaylist(ctx, queries.CreatePlaylistParams{
		ID:            id,
		Title:         title,
		Thumbnail:     "",
		Tracks:        make([]string, 0),
		AllowedTracks: make([]string, 0),
		Type:          queries.NullPlaylistType{PlaylistType: playlistType, Valid: true},
	}); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return dto2.Playlist{}, txErr
		}
		return dto2.Playlist{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return dto2.Playlist{}, txErr
		}
		return dto2.Playlist{}, err
	}

	return dto2.Playlist{
		Id:            id,
		Title:         title,
		Thumbnail:     "",
		Tracks:        nil,
		Length:        0,
		AllowedIds:    nil,
		AllowedLength: 0,
		Role:          queries.PlaylistRoleOwner,
		Type:          "tg", // TODO: SET
	}, nil
}

func (s *Playlist) GetById(ctx context.Context, playlistId string, userId int64) (dto2.Playlist, error) {
	q := queries.New(s.pool)
	playlist, err := q.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return dto2.Playlist{}, err
	}

	tracks := make([]dto2.Track, len(playlist.Tracks))
	for i, entity := range playlist.Tracks {
		dbTrack, err := q.GetTrackById(ctx, entity)
		if err != nil {
			return dto2.Playlist{}, err
		}

		tracks[i] = dto2.Track{
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

	return dto2.Playlist{
		Id:           playlist.ID,
		Title:        playlist.Title,
		Thumbnail:    playlist.Thumbnail,
		Tracks:       tracks,
		AllowedIds:   playlist.AllowedTracks,
		Count:        int(count),
		Length:       int(time),
		AllowedCount: int(allowedCount),
		Role:         playlist.Role,
		Type:         string(queries.PlaylistTypeUnknown),
	}, nil
}

func (s *Playlist) GetAll(ctx context.Context, userId int64) ([]dto2.Playlist, error) {
	q := queries.New(s.pool)
	playlists, err := q.GetUserPlaylists(ctx, userId)
	if err != nil {
		return nil, err
	}

	result := make([]dto2.Playlist, len(playlists))
	for i, playlist := range playlists {
		count := playlist.Count.Int32
		allowedCount := playlist.AllowedCount.Int32
		time := playlist.Time

		result[i] = dto2.Playlist{
			Id:           playlist.ID,
			Title:        playlist.Title,
			Thumbnail:    playlist.Thumbnail,
			Count:        int(count),
			Length:       int(time),
			AllowedCount: int(allowedCount),
			Tracks:       make([]dto2.Track, 0),
			AllowedIds:   make([]string, 0),
			Role:         playlist.Role,
			Type:         string(queries.PlaylistTypeUnknown),
		}
	}

	return result, nil
}

func (s *Playlist) Rename(ctx context.Context, playlistId, title string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	entity, err := q.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	entity.Title = title

	err = q.EditPlaylist(ctx, queries.EditPlaylistParams{
		ID:            entity.ID,
		Title:         entity.Title,
		Thumbnail:     entity.Thumbnail,
		Tracks:        entity.Tracks,
		AllowedTracks: entity.AllowedTracks,
		Type:          entity.Type,
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

func (s *Playlist) UpdatePhoto(ctx context.Context, playlistId, thumbnail string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	playlist, err := q.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	playlist.Thumbnail = thumbnail

	err = q.EditPlaylist(ctx, queries.EditPlaylistParams{
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

func (s *Playlist) Delete(ctx context.Context, playlistId string) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	err := q.DeletePlaylist(ctx, playlistId)
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
