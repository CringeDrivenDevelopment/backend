package service

import (
	"backend/internal/infra/queries"
	"backend/internal/transport/api/dto"
	"backend/pkg/utils"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

type Playlist struct {
	pool *pgxpool.Pool
}

func NewPlaylist(pool *pgxpool.Pool) *Playlist {
	return &Playlist{pool: pool}
}

func (s *Playlist) Create(ctx context.Context, title string, playlistType queries.PlaylistType, telegramId int64) (dto.Playlist, error) {
	id := ulid.Make().String()

	err := utils.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.CreatePlaylist(ctx, queries.CreatePlaylistParams{
			ID:            id,
			Title:         title,
			Thumbnail:     "",
			Tracks:        make([]string, 0),
			AllowedTracks: make([]string, 0),
			Type:          playlistType,
			ExternalID:    "",
			TelegramID:    telegramId,
		})
	})
	if err != nil {
		return dto.Playlist{}, err
	}

	return dto.Playlist{
		Id:    id,
		Title: title,
		Type:  string(playlistType),
	}, nil
}

func (s *Playlist) GetByGroup(ctx context.Context, telegramId int64) (dto.Playlist, error) {
	rq := queries.New(s.pool)
	playlist, err := rq.GetGroupPlaylist(ctx, telegramId)
	if err != nil {
		return dto.Playlist{}, err
	}

	tracks := make([]dto.Track, len(playlist.Tracks))
	for i, entity := range playlist.Tracks {
		dbTrack, err := rq.GetTrackById(ctx, entity)
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

	return dto.Playlist{
		Id:           playlist.ID,
		Title:        playlist.Title,
		Thumbnail:    playlist.Thumbnail,
		Tracks:       tracks,
		AllowedIds:   playlist.AllowedTracks,
		Count:        int(count),
		Length:       int(time),
		AllowedCount: int(allowedCount),
		Role:         "",
		Type:         string(playlist.Type),
	}, nil
}

func (s *Playlist) GetById(ctx context.Context, playlistId string, userId int64) (dto.Playlist, error) {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return dto.Playlist{}, err
	}

	tracks := make([]dto.Track, len(playlist.Tracks))
	for i, entity := range playlist.Tracks {
		dbTrack, err := rq.GetTrackById(ctx, entity)
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

	return dto.Playlist{
		Id:           playlist.ID,
		Title:        playlist.Title,
		Thumbnail:    playlist.Thumbnail,
		Tracks:       tracks,
		AllowedIds:   playlist.AllowedTracks,
		Count:        int(count),
		Length:       int(time),
		AllowedCount: int(allowedCount),
		Role:         playlist.Role,
		Type:         string(playlist.Type),
	}, nil
}

func (s *Playlist) GetAll(ctx context.Context, userId int64) ([]dto.Playlist, error) {
	rq := queries.New(s.pool)
	playlists, err := rq.GetUserPlaylists(ctx, userId)
	if err != nil {
		return nil, err
	}

	result := make([]dto.Playlist, len(playlists))
	for i, playlist := range playlists {
		count := playlist.Count.Int32
		allowedCount := playlist.AllowedCount.Int32
		time := playlist.Time

		result[i] = dto.Playlist{
			Id:           playlist.ID,
			Title:        playlist.Title,
			Thumbnail:    playlist.Thumbnail,
			Count:        int(count),
			Length:       int(time),
			AllowedCount: int(allowedCount),
			Tracks:       make([]dto.Track, 0),
			AllowedIds:   make([]string, 0),
			Role:         playlist.Role,
			Type:         string(playlist.Type),
		}
	}

	return result, nil
}

func (s *Playlist) Rename(ctx context.Context, playlistId, title string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	playlist.Title = title

	return utils.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditPlaylist(ctx, queries.EditPlaylistParams{
			ID:    playlist.ID,
			Title: playlist.Title,
		})
	})
}

func (s *Playlist) UpdatePhoto(ctx context.Context, playlistId, thumbnail string, userId int64) error {
	rq := queries.New(s.pool)
	playlist, err := rq.GetUserPlaylistById(ctx, queries.GetUserPlaylistByIdParams{
		PlaylistID: playlistId,
		UserID:     userId,
	})
	if err != nil {
		return err
	}

	playlist.Thumbnail = thumbnail

	return utils.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditPlaylist(ctx, queries.EditPlaylistParams{
			ID:        playlist.ID,
			Thumbnail: playlist.Thumbnail,
		})
	})
}

func (s *Playlist) Delete(ctx context.Context, playlistId string) error {
	return utils.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.DeletePlaylist(ctx, playlistId)
	})
}
