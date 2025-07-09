package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"net/url"
	"time"
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
		client: &http.Client{
			Timeout: time.Second * 10,
		},
		pool:      app.DB,
		baseUrl:   app.Settings.YoutubeUrl,
		authToken: app.Settings.YoutubeToken,
	}
}

func (s *YoutubeService) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.baseUrl+endpoint, body)
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
	resp, err := s.makeRequest(ctx, http.MethodGet, "/api/search?query="+url.QueryEscape(query), nil)
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
	trackTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	trackQueries := repository.New(trackTx)

	for i, track := range data {
		_, err := searchQueries.GetTrackById(ctx, track.Id)
		if errors.Is(err, pgx.ErrNoRows) {
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
		} else if err != nil {
			return nil, err
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

	err = trackTx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *YoutubeService) Download(ctx context.Context, id string) error {
	resp, err := s.makeRequest(ctx, http.MethodPost, "/api/dl?id="+id, nil)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	return nil
}

func (s *YoutubeService) Archive(ctx context.Context, songs []string) (dto.Archive, error) {
	bodyStruct := struct {
		Songs []string `json:"songs"`
	}{
		Songs: songs,
	}
	bodyBytes, err := sonic.Marshal(bodyStruct)
	if err != nil {
		return dto.Archive{}, err
	}
	resp, err := s.makeRequest(ctx, http.MethodPost, "/api/archive", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return dto.Archive{}, err
	}

	respStruct := dto.Archive{}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return dto.Archive{}, err
	}
	err = sonic.Unmarshal(respBytes, &respStruct)
	if err != nil {
		return dto.Archive{}, err
	}
	if err := resp.Body.Close(); err != nil {
		return dto.Archive{}, err
	}

	return respStruct, nil
}

func (s *YoutubeService) Stream(ctx context.Context, id *string, file string) (io.ReadCloser, error) {
	fileUrl := fmt.Sprintf("/api/dl/%s", file)
	if id != nil {
		fileUrl = fmt.Sprintf("/api/dl/%s/%s", *id, file)
	}

	resp, err := s.makeRequest(ctx, http.MethodGet, fileUrl, nil)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
