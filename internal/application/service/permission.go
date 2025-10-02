package service

import (
	"backend/internal/application"
	"backend/internal/domain/models"
	queries2 "backend/internal/domain/queries"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Permission struct {
	pool *pgxpool.Pool
}

func NewPermissionService(app *application.App) *Permission {
	return &Permission{pool: app.DB}
}

func (s *Permission) Add(ctx context.Context, role queries2.PlaylistRole, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries2.New(tx)
	rq := queries2.New(s.pool)
	_, err := rq.GetUserById(ctx, userId)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		err = q.CreateUser(ctx, userId)
		if err != nil {
			txErr := tx.Rollback(ctx)
			if txErr != nil {
				return txErr
			}
			return err
		}
	}
	err = q.CreateRole(ctx, queries2.CreateRoleParams{
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

func (s *Permission) AddGroup(ctx context.Context, playlist string, users []models.ParticipantData) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries2.New(tx)
	rq := queries2.New(s.pool)
	for _, user := range users {
		_, err := rq.GetUserById(ctx, user.UserID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			err = q.CreateUser(ctx, user.UserID)
			if err != nil {
				txErr := tx.Rollback(ctx)
				if txErr != nil {
					return txErr
				}
				return err
			}
		}

		err = q.CreateRole(ctx, queries2.CreateRoleParams{
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

func (s *Permission) Remove(ctx context.Context, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries2.New(tx)
	err := q.DeleteRole(ctx, queries2.DeleteRoleParams{
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

func (s *Permission) Edit(ctx context.Context, role queries2.PlaylistRole, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries2.New(tx)
	err := q.EditRole(ctx, queries2.EditRoleParams{
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

func (s *Permission) Get(ctx context.Context, userId int64, role queries2.PlaylistRole) (string, error) {
	q := queries2.New(s.pool)
	playlistId, err := q.GetRole(ctx, queries2.GetRoleParams{
		Role:   role,
		UserID: userId,
	})
	if err != nil {
		return "", err
	}

	return playlistId, nil
}
