package service

import (
	"backend/internal/application"

	"gopkg.in/telebot.v4"
)

type TgNotification struct {
	client *telebot.Bot
}

func NewNotificationService(app *application.App) *TgNotification {
	token := app.Settings.BotToken
	if app.Settings.Debug {
		token += "/test"
	}

	bot, err := telebot.NewBot(telebot.Settings{
		Token: token,
	})

	if err != nil {
		app.Logger.Fatal(err.Error())
		return nil
	}

	return &TgNotification{
		client: bot,
	}
}

func (s *TgNotification) Send(chatID int64, text string) error {
	_, err := s.client.Send(&telebot.Chat{
		ID: chatID,
	}, text)
	if err != nil {
		return err
	}

	return nil
}
