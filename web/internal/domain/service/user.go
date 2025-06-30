package service

import (
	"backend/internal/adapters/repository"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	pool *pgxpool.Pool
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{pool: pool}
}

func (s *UserService) Create(ctx context.Context, params repository.CreateUserParams) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	queries := repository.New(tx)

	if _, err := queries.GetUserById(ctx, params.ID); err == nil {
		return errors.New("user already exists")
	}

	if err := queries.CreateUser(ctx, params); err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	return tx.Commit(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (repository.User, error) {
	queries := repository.New(s.pool)

	user, err := queries.GetUserById(ctx, id)
	if err != nil {
		return repository.User{}, err
	}

	return user, nil
}
