package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"backend/internal/domain/utils"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionService struct {
	pool *pgxpool.Pool
}

func NewPermissionService(app *app.App) *PermissionService {
	return &PermissionService{
		pool: app.DB,
	}
}

func (s *PermissionService) Add(ctx context.Context, role, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	searchQueries := repository.New(s.pool)
	_, err := searchQueries.GetUserById(ctx, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		err = queries.CreateUser(ctx, repository.CreateUserParams{
			ID:   userId,
			Name: fmt.Sprintf("user%d", userId),
		})
		if err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				return txErr
			}
			return err
		}
	}
	err = queries.CreateRole(ctx, repository.CreateRoleParams{
		Role:       role,
		UserID:     userId,
		PlaylistID: playlist,
	})
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}

func (s *PermissionService) AddGroup(ctx context.Context, playlist string, users []utils.ParticipantData) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	searchQueries := repository.New(s.pool)
	for _, user := range users {
		_, err := searchQueries.GetUserById(ctx, user.UserID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			err = queries.CreateUser(ctx, repository.CreateUserParams{
				ID:   user.UserID,
				Name: fmt.Sprintf("user%d", user.UserID),
			})
			if err != nil {
				txErr := tx.Rollback(ctx)
				if txErr != nil {
					return txErr
				}
				return err
			}
		}

		err = queries.CreateRole(ctx, repository.CreateRoleParams{
			Role:       user.NewRole,
			UserID:     user.UserID,
			PlaylistID: playlist,
		})
		if err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				return txErr
			}
			return err
		}
	}

	err := tx.Commit(ctx)
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}

func (s *PermissionService) Remove(ctx context.Context, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	err := queries.DeleteRole(ctx, repository.DeleteRoleParams{
		PlaylistID: playlist,
		UserID:     userId,
	})
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}

func (s *PermissionService) Edit(ctx context.Context, role, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	err := queries.EditRole(ctx, repository.EditRoleParams{
		Role:       role,
		PlaylistID: playlist,
		UserID:     userId,
	})

	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			return txErr
		}
		return err
	}

	return nil
}

func (s *PermissionService) Get(ctx context.Context, userId int64, role string) (string, error) {
	queries := repository.New(s.pool)
	playlistId, err := queries.GetRole(ctx, repository.GetRoleParams{
		Role:   role,
		UserID: userId,
	})
	if err != nil {
		return "", err
	}

	return playlistId, nil
}
