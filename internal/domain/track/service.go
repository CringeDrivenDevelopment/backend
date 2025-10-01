package track

import (
	"backend/cmd/app"
	"backend/internal/domain/youtube"
	"backend/internal/infra/database/queries"
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool

	youtube *youtube.Service
}

func NewService(app *app.App) *Service {
	return &Service{pool: app.DB, youtube: youtube.NewService()}
}

func (s *Service) Search(ctx context.Context, query string, userId int64) ([]DtoTrack, error) {
	data, err := s.youtube.Search(query, youtube.FILTER_SONGS)
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

	tracks := make([]DtoTrack, len(results))
	for i, result := range results {
		track, err := youtube.ParseRaw(&result.Data)
		if err != nil {
			return nil, errors.New("could not parse music song")
		}

		tracks[i] = DtoTrack{
			Id:          track.Id,
			Title:       track.Title,
			Authors:     track.Authors,
			Thumbnail:   track.Thumbnail,
			Length:      int32(track.Length),
			Explicit:    track.Explicit,
			PlaylistIds: make([]string, 0),
		}
	}

	rq := queries.New(s.pool)
	trackTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	q := queries.New(trackTx)

	for i, track := range tracks {
		_, err := rq.GetTrackById(ctx, track.Id)
		if errors.Is(err, pgx.ErrNoRows) {
			err = q.CreateTrack(ctx, queries.CreateTrackParams{
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

		playlistIds, err := rq.GetTrackPlaylists(ctx, queries.GetTrackPlaylistsParams{
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

func (s *Service) GetById(ctx context.Context, id string) (DtoTrack, error) {
	q := queries.New(s.pool)
	track, err := q.GetTrackById(ctx, id)
	if err != nil {
		return DtoTrack{}, err
	}

	return DtoTrack{
		Id:        track.ID,
		Title:     track.Title,
		Authors:   track.Authors,
		Thumbnail: track.Thumbnail,
		Length:    track.Length,
		Explicit:  track.Explicit,
	}, nil
}

func (s *Service) Approve(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
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
	q := queries.New(tx)

	err = q.EditPlaylist(ctx, queries.EditPlaylistParams{
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

func (s *Service) Decline(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
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
	q := queries.New(tx)

	err = q.EditPlaylist(ctx, queries.EditPlaylistParams{
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

func (s *Service) Submit(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
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

	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}
	q := queries.New(tx)

	err = q.EditPlaylist(ctx, queries.EditPlaylistParams{
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

func (s *Service) Unapprove(ctx context.Context, playlistId, trackId string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetPlaylistById(ctx, queries.GetPlaylistByIdParams{
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
	q := queries.New(tx)

	err = q.EditPlaylist(ctx, queries.EditPlaylistParams{
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
