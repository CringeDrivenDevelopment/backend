package permission

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

func (s *Service) Add(ctx context.Context, role queries.PlaylistRole, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	rq := queries.New(s.pool)
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
	err = q.CreateRole(ctx, queries.CreateRoleParams{
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

func (s *Service) AddGroup(ctx context.Context, playlist string, users []ParticipantData) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	rq := queries.New(s.pool)
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

		err = q.CreateRole(ctx, queries.CreateRoleParams{
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

func (s *Service) Remove(ctx context.Context, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	err := q.DeleteRole(ctx, queries.DeleteRoleParams{
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

func (s *Service) Edit(ctx context.Context, role queries.PlaylistRole, playlist string, userId int64) error {
	tx, txErr := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if txErr != nil {
		return txErr
	}

	q := queries.New(tx)
	err := q.EditRole(ctx, queries.EditRoleParams{
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

func (s *Service) Get(ctx context.Context, userId int64, role queries.PlaylistRole) (string, error) {
	q := queries.New(s.pool)
	playlistId, err := q.GetRole(ctx, queries.GetRoleParams{
		Role:   role,
		UserID: userId,
	})
	if err != nil {
		return "", err
	}

	return playlistId, nil
}
