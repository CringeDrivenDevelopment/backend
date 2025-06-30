package utils

import (
	"backend/internal/adapters/repository"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/multierr"
	"time"
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

	/*
		initTx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			pool.Close()
			return nil, err
		}

		queries := repository.New(initTx)
	*/
	queries := repository.New(pool)
	err = multierr.Combine(
		queries.InitUsers(ctx),
		queries.InitTracks(ctx),
		queries.InitPlaylists(ctx),
		queries.InitRoleEnum(ctx),
		queries.InitPermissions(ctx),
	)
	if err != nil {
		//
		pool.Close()
		/*
			txErr := initTx.Rollback(ctx)
			if txErr != nil {
				return nil, txErr
			}
		*/
		return nil, err
	}

	/*
		err = initTx.Commit(ctx)
			if err != nil {
				pool.Close()
				return nil, err
			}
	*/

	return pool, nil
}
