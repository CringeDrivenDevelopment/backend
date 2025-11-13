package infra

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger(lc fx.Lifecycle) (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("logger initialized")

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("logger stopped")
			return logger.Sync()
		},
	})

	return logger, nil
}
