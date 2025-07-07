package service

import (
	"backend/cmd/app"
	"backend/internal/adapters/repository"
	"context"
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

func (s *PlaylistService) Add(ctx context.Context, role, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	queries := repository.New(tx)
	err := queries.CreateRole(ctx, repository.CreateRoleParams{
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

func (s *PlaylistService) Remove(ctx context.Context, playlist string, userId int64) error {
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

func (s *PlaylistService) Edit(ctx context.Context, role, playlist string, userId int64) error {
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
