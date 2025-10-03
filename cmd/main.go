package main

import (
	"backend/internal/application"
	"backend/internal/infra/handlers/api"
	"backend/internal/infra/handlers/bot"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("logger initialized")

	mainApp, err := application.New(logger)
	if err != nil {
		logger.Panic(err.Error())
		return
	}

	api.Setup(mainApp)

	logger.Info("endpoints mapped")

	botApp, err := bot.New(mainApp)
	if err != nil {
		logger.Panic(err.Error())
		return
	}

	botApp.Setup()

	logger.Info("app initialized")

	go func() {
		err := botApp.Start()
		if err != nil {
			logger.Info(err.Error())
		}
	}()

	go func() {
		err := mainApp.Start()
		if err != nil {
			logger.Info(err.Error())
		}

	}()

	<-sigChan

	botApp.Stop()
	err = mainApp.Stop()
	if err != nil {
		logger.Panic(err.Error())
	}
	mainApp.DB.Close()

	logger.Info("server stopped")
}
