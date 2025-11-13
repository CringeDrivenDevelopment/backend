package infra

import (
	projectroot "backend"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewPostgresConnection(lc fx.Lifecycle, logger *zap.Logger, cfg Config) (*pgxpool.Pool, error) {
	ctxWithCancel, cancel := context.WithCancel(context.Background())

	pool, err := pgxpool.New(ctxWithCancel, cfg.DbUrl)
	if err != nil {
		cancel()
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("running migrations with db url: " + cfg.DbUrl)

			// configure pool
			poolConfig := pool.Config()
			poolConfig.MaxConns = 10
			poolConfig.MinConns = 2
			poolConfig.MaxConnLifetime = time.Hour
			poolConfig.MaxConnIdleTime = time.Minute * 30
			poolConfig.HealthCheckPeriod = time.Minute

			// check if online
			if err := pool.Ping(ctx); err != nil {
				return err
			}

			// run migrations
			goose.SetBaseFS(projectroot.EmbedMigrations)
			if err := goose.SetDialect("postgres"); err != nil {
				return err
			}
			db := stdlib.OpenDBFromPool(pool)
			if err := goose.Up(db, "sql/migrations"); err != nil {
				return err
			}
			if err := db.Close(); err != nil {
				return err
			}

			logger.Info("migrations applied")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			cancel()

			logger.Info("db connection closed")

			return nil
		},
	})

	return pool, nil
}
