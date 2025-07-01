package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/dto"
	"context"
	"errors"
	"fmt"
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

func NewYoutubeService(app *app.App) *YoutubeService {
	return &YoutubeService{
		client:    &http.Client{},
		pool:      app.DB,
		baseUrl:   app.Settings.YoutubeUrl,
		authToken: app.Settings.YoutubeToken,
	}
}

func (s *YoutubeService) Search(ctx context.Context, query string) ([]dto.Track, error) {
	req, err := http.NewRequest(http.MethodGet, s.baseUrl+"/search?query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.authToken))

	resp, err := s.client.Do(req)
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
		return nil, err
	}

	searchQueries := repository.New(s.pool)

	for _, track := range data {
		if _, err := searchQueries.GetTrackById(ctx, track.Id); errors.Is(pgx.ErrNoRows, err) {
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
	}

	return data, nil
}
