package user

import (
	"backend/cmd/app"
	"backend/internal/infra/database/queries"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(app *app.App) *Service {
	return &Service{pool: app.DB}
}

func (s *Service) Create(ctx context.Context, id int64) error {
	readQueries := queries.New(s.pool)
	if _, err := readQueries.GetUserById(ctx, id); err == nil {
		return errors.New("user already exists")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	q := queries.New(tx)

	if err := q.CreateUser(ctx, id); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}

func (s *Service) GetByID(ctx context.Context, id int64) error {
	a := queries.New(s.pool)

	_, err := a.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
