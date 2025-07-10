package main

import (
	"backend/cmd/app"
	"backend/internal/adapters/handlers/api"
	"backend/internal/adapters/handlers/bot"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("logger initialized")

	mainApp, err := app.New(logger)
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
