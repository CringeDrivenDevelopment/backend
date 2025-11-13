package service

import (
	"backend/internal/app"
	"backend/internal/db"
	"backend/internal/db/queries"
	"backend/internal/errorz"
	"backend/internal/transport/api/dto"
	"backend/pkg/youtube"
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchAPI interface {
	Search(ctx context.Context, query string) ([]dto.Track, error)
}

type Track struct {
	pool *pgxpool.Pool

	youtube SearchAPI
	// spotify SearchAPI
}

func NewTrackService(app *app.App) *Track {
	return &Track{pool: app.DB, youtube: youtube.NewService()}
}

/*
Search - метод для поиска треков на какой-либо из площадок

Старая сигнатура:
Search(ctx context.Context, query string, userId int64) ([]dto.Track, error)
Изменилась, из-за поддержки множества площадок, а так же, из-за ненадобности получения списка плейлистов для трека
*/
func (s *Track) Search(ctx context.Context, query string) ([]dto.Track, error) {
	tracks, err := s.youtube.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	rq := queries.New(s.pool)

	if err := db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		for _, track := range tracks {
			_, err := rq.GetTrackById(ctx, track.Id)
			if errors.Is(err, pgx.ErrNoRows) {
				err = tq.CreateTrack(ctx, queries.CreateTrackParams{
					ID:        track.Id,
					Title:     track.Title,
					Authors:   track.Authors,
					Thumbnail: track.Thumbnail,
					Length:    track.Length,
					Explicit:  track.Explicit,
				})
				if err != nil {
					return err
				}
			}

			if err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (s *Track) GetById(ctx context.Context, id string) (dto.Track, error) {
	q := queries.New(s.pool)
	track, err := q.GetTrackById(ctx, id)
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
	}, nil
}

func (s *Track) Approve(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if playlist.Role != queries.PlaylistRoleOwner && playlist.Role != queries.PlaylistRoleModerator {
		return errorz.ErrNotEnoughPerms
	}

	if slices.Contains(playlist.AllowedTracks, trackId) {
		return nil
	}

	if !slices.Contains(playlist.Tracks, trackId) {
		return pgx.ErrNoRows
	}

	if err := db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditPlaylist(ctx, queries.EditPlaylistParams{
			ID:            playlistId,
			AllowedTracks: append(playlist.AllowedTracks, trackId),
		})
	}); err != nil {
		return err
	}

	return nil
}

func (s *Track) Decline(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if playlist.Role != queries.PlaylistRoleOwner && playlist.Role != queries.PlaylistRoleModerator {
		return errorz.ErrNotEnoughPerms
	}

	if slices.Contains(playlist.AllowedTracks, trackId) || !slices.Contains(playlist.Tracks, trackId) {
		return pgx.ErrNoRows
	}

	for i, track := range playlist.Tracks {
		if track == trackId {
			playlist.Tracks = append(playlist.Tracks[:i], playlist.Tracks[i+1:]...)
			break
		}
	}

	if err := db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditPlaylist(ctx, queries.EditPlaylistParams{
			ID:            playlistId,
			AllowedTracks: playlist.AllowedTracks,
		})
	}); err != nil {
		return err
	}

	return nil
}

func (s *Track) Submit(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if _, err := rq.GetTrackById(ctx, trackId); err != nil {
		return err
	}

	tracks := playlist.Tracks
	allowedTracks := playlist.AllowedTracks
	if (playlist.Role == queries.PlaylistRoleOwner || playlist.Role == queries.PlaylistRoleModerator) && !slices.Contains(allowedTracks, trackId) {
		tracks = append(tracks, trackId)
		allowedTracks = append(allowedTracks, trackId)
	} else if !slices.Contains(tracks, trackId) && playlist.Role == queries.PlaylistRoleViewer {
		tracks = append(tracks, trackId)
	} else {
		return nil
	}

	if err := db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditPlaylist(ctx, queries.EditPlaylistParams{
			ID:            playlistId,
			Tracks:        tracks,
			AllowedTracks: allowedTracks,
		})
	}); err != nil {
		return err
	}

	return nil
}

func (s *Track) Unapprove(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	if playlist.Role != queries.PlaylistRoleOwner && playlist.Role != queries.PlaylistRoleModerator {
		return errorz.ErrNotEnoughPerms
	}

	if !slices.Contains(playlist.AllowedTracks, trackId) || !slices.Contains(playlist.Tracks, trackId) {
		return pgx.ErrNoRows
	}

	for i, track := range playlist.AllowedTracks {
		if track == trackId {
			playlist.AllowedTracks = append(playlist.AllowedTracks[:i], playlist.AllowedTracks[i+1:]...)
			break
		}
	}

	if err := db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditPlaylist(ctx, queries.EditPlaylistParams{
			ID:            playlistId,
			AllowedTracks: playlist.AllowedTracks,
		})
	}); err != nil {
		return err
	}

	return nil
}
