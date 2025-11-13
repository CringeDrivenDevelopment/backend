package service

import (
	"backend/internal/app"
	"backend/internal/db"
	"backend/internal/db/queries"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	pool *pgxpool.Pool
}

func NewUserService(app *app.App) *User {
	return &User{pool: app.DB}
}

func (s *User) Create(ctx context.Context, id int64) error {
	readQueries := queries.New(s.pool)
	if _, err := readQueries.GetUserById(ctx, id); err == nil {
		return errors.New("user already exists")
	}

	if err := db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.CreateUser(ctx, id)
	}); err != nil {
		return err
	}

	return nil
}

func (s *User) GetByID(ctx context.Context, id int64) error {
	a := queries.New(s.pool)

	_, err := a.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
