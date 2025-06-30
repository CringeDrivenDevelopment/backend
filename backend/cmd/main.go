package main

import (
	"backend/cmd/app"
	"backend/internal/adapters/controller"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("logger initialized")

	mainApp, err := app.New(logger)
	if err != nil {
		logger.Panic(err.Error())
	}

	if mainApp == nil {
		logger.Panic("mainApp is nil")
	}

	logger.Info("app initialized")

	controller.Setup(mainApp)

	logger.Info("endpoints mapped")
	err = mainApp.Start()
	if err != nil {
		logger.Panic(err.Error())
	}
}
