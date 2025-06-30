package main

import (
	"context"
	"fmt"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/tg"
	"log"
)

func main() {
	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		-52,
		// Get ApiHash from https://my.telegram.org/apps
		"52_Bro",
		// ClientType, as we defined above
		gotgproto.ClientTypeBot("teeeeegeeeetokeeeen"),
		// Optional parameters of client
		&gotgproto.ClientOpts{
			Session: sessionMaker.SqlSession(sqlite.Open("bot.db")),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dispatcher := client.Dispatcher

	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))

	// Set menu button on startup
	setMenuButton(client)

	fmt.Printf("client (@%s) has been started...\n", client.Self.Username)

	client.Idle()
}

func setMenuButton(client *gotgproto.Client) {
	// Create mini-app button with query parameters
	button := &tg.BotMenuButton{
		Text: "Открыть MiniApp",
		URL:  "https://lxft.tech",
	}

	// Set as bot's menu button
	_, err := client.API().BotsSetBotMenuButton(context.Background(), &tg.BotsSetBotMenuButtonRequest{
		Button: button,
	})
	if err != nil {
		log.Println("WARNING: Failed to set menu button:", err)
	}
}

func startHandler(ctx *ext.Context, update *ext.Update) error {
	msg := update.EffectiveMessage

	// Build URL with query parameters
	webAppURL := "https://lxft.tech"

	// Create inline keyboard
	btn := tg.KeyboardButtonWebView{
		Text: "Открыть MiniApp",
		URL:  webAppURL,
	}
	row := tg.KeyboardButtonRow{Buttons: []tg.KeyboardButtonClass{&btn}}
	markup := tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{row}}

	// Send message with button
	_, err := ctx.Reply(update, ext.ReplyTextString("Привет! Чтобы начать работу, кликни ниже:"), &ext.ReplyOpts{
		NoWebpage:        false,
		Markup:           &markup,
		ReplyToMessageId: msg.ID,
	})
	return err
}
