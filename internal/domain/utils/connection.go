package utils

import (
	"backend/internal/adapters/repository"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/multierr"
)

func NewConnection(ctx context.Context, connUrl string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connUrl)
	if err != nil {
		return nil, err
	}

	config := pool.Config()
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	queries := repository.New(pool)
	err = multierr.Combine(
		queries.InitUsers(ctx),
		queries.InitTracks(ctx),
		queries.InitPlaylists(ctx),
		queries.InitPermissions(ctx),
		queries.InitTracksIndex(ctx),
		queries.InitPermissionsIndex(ctx),
		queries.InitPlaylistsTracksIndex(ctx),
		queries.InitPlaylistsAllowedTracksIndex(ctx),
		queries.InitCalculatePlaylistTime(ctx),
		queries.InitUpdatePlaylistTimes(ctx),
		queries.InitPlaylistTimesTrigger(ctx),
		queries.InitUpdatePlaylistOnTrackChange(ctx),
		queries.InitTrackUpdate(ctx),
	)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
