package database

import (
	projectroot "backend"
	"backend/internal/infra/database/queries"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
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

	goose.SetBaseFS(projectroot.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		pool.Close()
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "sql/migrations"); err != nil {
		panic(err)
	}
	if err := db.Close(); err != nil {
		panic(err)
	}

	return pool, nil
}

func NewQuery(pool *pgxpool.Pool) *queries.Queries {
	return queries.New(pool)
}

func NewTxQuery(ctx context.Context, pool *pgxpool.Pool) (*queries.Queries, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return queries.New(tx), nil
}
