package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"context"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"net/url"
)

type YoutubeService struct {
	client *http.Client
	pool   *pgxpool.Pool

	baseUrl   string
	authToken string
}

type searchError struct {
	Error string `json:"error"`
}

func NewYoutubeService(app *app.App) *YoutubeService {
	return &YoutubeService{
		client:    &http.Client{},
		pool:      app.DB,
		baseUrl:   app.Settings.YoutubeUrl,
		authToken: app.Settings.YoutubeToken,
	}
}

func (s *YoutubeService) makeRequest(ctx context.Context, method string, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.baseUrl+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return resp, nil
}

func (s *YoutubeService) Search(ctx context.Context, query string, userId int64) ([]dto.Track, error) {
	resp, err := s.makeRequest(ctx, http.MethodGet, "/api/search?query="+url.QueryEscape(query))
	if err != nil {
		return nil, err
	}

	var data []dto.Track
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	err = sonic.Unmarshal(body, &data)
	if err != nil {
		var apiErr searchError
		sonicErr := sonic.Unmarshal(body, &apiErr)
		if sonicErr != nil {
			return nil, sonicErr
		}

		return nil, errors.New(apiErr.Error)
	}

	searchQueries := repository.New(s.pool)

	for i, track := range data {
		if _, err := searchQueries.GetTrackById(ctx, track.Id); errors.Is(err, pgx.ErrNoRows) {
			trackTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				return nil, err
			}
			trackQueries := repository.New(trackTx)
			err = trackQueries.CreateTrack(ctx, repository.CreateTrackParams{
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
			err = trackTx.Commit(ctx)
			if err != nil {
				return nil, err
			}
		}

		playlistIds, err := searchQueries.GetTrackPlaylists(ctx, repository.GetTrackPlaylistsParams{
			TrackID: track.Id,
			UserID:  userId,
		})
		if err != nil {
			track.PlaylistIds = make([]string, 0)
		} else {
			data[i].PlaylistIds = playlistIds
		}
	}

	return data, nil
}

func (s *YoutubeService) Download(ctx context.Context, id string) error {
	resp, err := s.makeRequest(ctx, http.MethodPost, "/api/dl?id="+id)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	return nil
}
