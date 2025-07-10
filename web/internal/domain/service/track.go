package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"slices"
)

type TrackService struct {
	pool *pgxpool.Pool
}

func NewTrackService(app *app.App) *TrackService {
	return &TrackService{pool: app.DB}
}

func (s *TrackService) GetById(ctx context.Context, id string) (dto.Track, error) {
	queries := repository.New(s.pool)
	track, err := queries.GetTrackById(ctx, id)
	if err != nil {
		return dto.Track{}, err
	}

	return dto.Track{
		Id:        track.ID,
		Title:     track.Title,
		Authors:   track.Authors,
		Thumbnail: track.Thumbnail,
		Length:    track.Length,
		Explicit:  track.Explicit,
	}, err
}

func (s *TrackService) Approve(ctx context.Context, playlistId, trackId string, userId int64) error {
	readQueries := repository.New(s.pool)
	playlist, err := readQueries.GetPlaylistById(ctx, repository.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if slices.Contains(playlist.AllowedTracks, trackId) {
		return nil
	}

	if !slices.Contains(playlist.Tracks, trackId) {
		return errors.New("track not found in playlist")
	}

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	queries := repository.New(tx)

	err = queries.EditPlaylist(ctx, repository.EditPlaylistParams{
		ID:            playlistId,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        playlist.Tracks,
		AllowedTracks: append(playlist.AllowedTracks, trackId),
		Type:          playlist.Type,
	})
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
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

func (s *TrackService) Decline(ctx context.Context, playlistId, trackId string, userId int64) error {
	readQueries := repository.New(s.pool)
	playlist, err := readQueries.GetPlaylistById(ctx, repository.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if slices.Contains(playlist.AllowedTracks, trackId) {
		return errors.New("track is allowed")
	}

	if !slices.Contains(playlist.Tracks, trackId) {
		return errors.New("track not found in playlist")
	}

	for i, track := range playlist.Tracks {
		if track == trackId {
			playlist.Tracks = append(playlist.Tracks[:i], playlist.Tracks[i+1:]...)
			break
		}
	}

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	queries := repository.New(tx)

	err = queries.EditPlaylist(ctx, repository.EditPlaylistParams{
		ID:            playlistId,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        playlist.Tracks,
		AllowedTracks: playlist.AllowedTracks,
		Type:          playlist.Type,
	})
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
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

func (s *TrackService) Submit(ctx context.Context, playlistId, trackId string, userId int64) error {
	readQueries := repository.New(s.pool)
	playlist, err := readQueries.GetPlaylistById(ctx, repository.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if _, err := readQueries.GetTrackById(ctx, trackId); err != nil {
		return err
	}

	tracks := playlist.Tracks
	allowedTracks := playlist.AllowedTracks
	if (playlist.Role == dto.OwnerRole || playlist.Role == dto.ModeratorRole) && !slices.Contains(allowedTracks, trackId) {
		tracks = append(tracks, trackId)
		allowedTracks = append(allowedTracks, trackId)
	} else if !slices.Contains(tracks, trackId) && playlist.Role == dto.ViewerRole {
		tracks = append(tracks, trackId)
	} else {
		return nil
	}

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	queries := repository.New(tx)

	err = queries.EditPlaylist(ctx, repository.EditPlaylistParams{
		ID:            playlistId,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        tracks,
		AllowedTracks: allowedTracks,
		Type:          playlist.Type,
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

func (s *TrackService) RemoveApproved(ctx context.Context, playlistId, trackId string, userId int64) error {
	readQueries := repository.New(s.pool)
	playlist, err := readQueries.GetPlaylistById(ctx, repository.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if !slices.Contains(playlist.AllowedTracks, trackId) {
		return errors.New("track not found in allowed")
	}

	if !slices.Contains(playlist.Tracks, trackId) {
		return errors.New("track not found in playlist")
	}

	for i, track := range playlist.AllowedTracks {
		if track == trackId {
			playlist.AllowedTracks = append(playlist.AllowedTracks[:i], playlist.AllowedTracks[i+1:]...)
			break
		}
	}

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	queries := repository.New(tx)

	err = queries.EditPlaylist(ctx, repository.EditPlaylistParams{
		ID:            playlistId,
		Title:         playlist.Title,
		Thumbnail:     playlist.Thumbnail,
		Tracks:        playlist.Tracks,
		AllowedTracks: playlist.AllowedTracks,
		Type:          playlist.Type,
	})
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
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
