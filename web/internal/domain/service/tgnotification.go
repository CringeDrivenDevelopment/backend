package service

import (
	"backend/cmd/app"
	"gopkg.in/telebot.v4"
)

type TelegramNotificationService struct {
	client *telebot.Bot
}

func NewTelegramNotificationService(app *app.App) *TelegramNotificationService {
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

	return &TelegramNotificationService{
		client: bot,
	}
}

func (s *TelegramNotificationService) Send(chatID int64, text string) error {
	_, err := s.client.Send(&telebot.Chat{
		ID: chatID,
	}, text)
	if err != nil {
		return err
	}

	return nil
}
