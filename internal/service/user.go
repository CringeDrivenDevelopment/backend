package service

import (
	"backend/internal/infra/queries"
	"backend/pkg/utils"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	pool *pgxpool.Pool
}

func NewUser(pool *pgxpool.Pool) *User {
	return &User{pool: pool}
}

func (s *User) Create(ctx context.Context, id int64) error {
	rq := queries.New(s.pool)
	if _, err := rq.GetUserById(ctx, id); err == nil {
		return nil
	}

	if err := utils.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.CreateUser(ctx, id)
	}); err != nil {
		return err
	}

	return nil
}

func (s *User) GetByID(ctx context.Context, id int64) error {
	rq := queries.New(s.pool)

	_, err := rq.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
