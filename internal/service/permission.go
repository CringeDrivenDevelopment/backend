package service

import (
	"backend/internal/app"
	"backend/internal/db"
	"backend/internal/db/queries"
	"backend/internal/transport/bot/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Permission struct {
	pool *pgxpool.Pool
}

func NewPermissionService(app *app.App) *Permission {
	return &Permission{pool: app.DB}
}

func (s *Permission) Add(ctx context.Context, role queries.PlaylistRole, playlist string, userId int64) error {
	rq := queries.New(s.pool)

	return db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		_, err := rq.GetUserById(ctx, userId)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}

			if err := tq.CreateUser(ctx, userId); err != nil {
				return err
			}
		}

		return tq.CreateRole(ctx, queries.CreateRoleParams{
			Role:       role,
			UserID:     userId,
			PlaylistID: playlist,
		})
	})
}

func (s *Permission) AddGroup(ctx context.Context, playlist string, users []models.ParticipantData) error {
	rq := queries.New(s.pool)

	return db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		for _, user := range users {
			_, err := rq.GetUserById(ctx, user.UserID)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return err
				}
				err = tq.CreateUser(ctx, user.UserID)
				if err != nil {
					return err
				}
			}

			err = tq.CreateRole(ctx, queries.CreateRoleParams{
				Role:       user.NewRole,
				UserID:     user.UserID,
				PlaylistID: playlist,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Permission) Remove(ctx context.Context, playlist string, userId int64) error {
	return db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.DeleteRole(ctx, queries.DeleteRoleParams{
			PlaylistID: playlist,
			UserID:     userId,
		})
	})
}

func (s *Permission) Edit(ctx context.Context, role queries.PlaylistRole, playlist string, userId int64) error {
	return db.ExecInTx(ctx, s.pool, func(tq *queries.Queries) error {
		return tq.EditRole(ctx, queries.EditRoleParams{
			Role:       role,
			PlaylistID: playlist,
			UserID:     userId,
		})
	})
}

func (s *Permission) Get(ctx context.Context, userId int64, role queries.PlaylistRole) (string, error) {
	q := queries.New(s.pool)

	return q.GetRole(ctx, queries.GetRoleParams{
		Role:   role,
		UserID: userId,
	})
}
