package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool *pgxpool.Pool
}

func NewUserService(app *app.App) *UserService {
	return &UserService{pool: app.DB}
}

func (s *UserService) Create(ctx context.Context, params repository.CreateUserParams) error {
	readQueries := repository.New(s.pool)
	if _, err := readQueries.GetUserById(ctx, params.ID); err == nil {
		return errors.New("user already exists")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	queries := repository.New(tx)

	if err := queries.CreateUser(ctx, params); err != nil {
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

func (s *UserService) GetByID(ctx context.Context, id int64) (repository.User, error) {
	queries := repository.New(s.pool)

	user, err := queries.GetUserById(ctx, id)
	if err != nil {
		return repository.User{}, err
	}

	return user, nil
}
