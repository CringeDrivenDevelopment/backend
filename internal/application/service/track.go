package service

import (
	"backend/internal/application"
	"backend/internal/domain/models"
	queries2 "backend/internal/domain/queries"
	"backend/internal/infra/youtube"
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Track struct {
	pool *pgxpool.Pool

	ytApi *youtube.Service
}

func NewTrackService(app *application.App) *Track {
	return &Track{pool: app.DB, ytApi: youtube.NewService()}
}

func (s *Track) Search(ctx context.Context, query string, userId int64) ([]models.DtoTrack, error) {
	data, err := s.ytApi.Search(query, youtube.FILTER_SONGS)
	if err != nil {
		return nil, err
	}

	var results []struct {
		Data youtube.RawYtMusicSong `json:"musicResponsiveListItemRenderer"`
	}

	for _, tab := range data.Contents.TabbedSearchResultsRenderer.Tabs {
		for _, content := range tab.TabRenderer.Content.SectionListRenderer.Contents {
			if content.MusicShelfRenderer.Contents == nil {
				continue
			}

			results = *content.MusicShelfRenderer.Contents
			break
		}
	}

	if results == nil {
		return nil, errors.New("could not find music shelf")
	}

	tracks := make([]models.DtoTrack, len(results))
	for i, result := range results {
		track, err := youtube.ParseRaw(&result.Data)
		if err != nil {
			return nil, errors.New("could not parse music song")
		}

		tracks[i] = models.DtoTrack{
			Id:          track.Id,
			Title:       track.Title,
			Authors:     track.Authors,
			Thumbnail:   track.Thumbnail,
			Length:      int32(track.Length),
			Explicit:    track.Explicit,
			PlaylistIds: make([]string, 0),
		}
	}

	rq := queries2.New(s.pool)
	trackTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	q := queries2.New(trackTx)

	for i, track := range tracks {
		_, err := rq.GetTrackById(ctx, track.Id)
		if errors.Is(err, pgx.ErrNoRows) {
			err = q.CreateTrack(ctx, queries2.CreateTrackParams{
				ID:        track.Id,
				Title:     track.Title,
				Authors:   track.Authors,
				Thumbnail: track.Thumbnail,
				Length:    track.Length,
				Explicit:  track.Explicit,
			})
			if err != nil {
				if txErr := trackTx.Rollback(ctx); txErr != nil {
					return nil, err
				}
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		playlistIds, err := rq.GetTrackPlaylists(ctx, queries2.GetTrackPlaylistsParams{
			TrackID: track.Id,
			UserID:  userId,
		})
		if err != nil {
			track.PlaylistIds = make([]string, 0)
		} else {
			tracks[i].PlaylistIds = playlistIds
		}
	}

	err = trackTx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func (s *Track) GetById(ctx context.Context, id string) (models.DtoTrack, error) {
	q := queries2.New(s.pool)
	track, err := q.GetTrackById(ctx, id)
	if err != nil {
		return models.DtoTrack{}, err
	}

	return models.DtoTrack{
		Id:        track.ID,
		Title:     track.Title,
		Authors:   track.Authors,
		Thumbnail: track.Thumbnail,
		Length:    track.Length,
		Explicit:  track.Explicit,
	}, nil
}

func (s *Track) Approve(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries2.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries2.GetPlaylistByIdParams{
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
		return pgx.ErrNoRows
	}

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	q := queries2.New(tx)

	err = q.EditPlaylist(ctx, queries2.EditPlaylistParams{
		ID:            playlistId,
		AllowedTracks: append(playlist.AllowedTracks, trackId),
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

func (s *Track) Decline(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries2.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries2.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
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

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	q := queries2.New(tx)

	err = q.EditPlaylist(ctx, queries2.EditPlaylistParams{
		ID:            playlistId,
		AllowedTracks: playlist.AllowedTracks,
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

func (s *Track) Submit(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries2.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries2.GetPlaylistByIdParams{
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
	if (playlist.Role == queries2.PlaylistRoleOwner || playlist.Role == queries2.PlaylistRoleModerator) && !slices.Contains(allowedTracks, trackId) {
		tracks = append(tracks, trackId)
		allowedTracks = append(allowedTracks, trackId)
	} else if !slices.Contains(tracks, trackId) && playlist.Role == queries2.PlaylistRoleViewer {
		tracks = append(tracks, trackId)
	} else {
		return nil
	}

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	q := queries2.New(tx)

	err = q.EditPlaylist(ctx, queries2.EditPlaylistParams{
		ID:            playlistId,
		Tracks:        tracks,
		AllowedTracks: allowedTracks,
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

func (s *Track) Unapprove(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries2.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries2.GetPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
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

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	q := queries2.New(tx)

	err = q.EditPlaylist(ctx, queries2.EditPlaylistParams{
		ID:            playlistId,
		AllowedTracks: playlist.AllowedTracks,
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
