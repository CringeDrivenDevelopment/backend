package bot

import (
	"backend/cmd/app"
	"backend/internal/domain/service"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram/dcs"
	"go.uber.org/zap"
)

type Bot struct {
	playlistService   *service.PlaylistService
	permissionService *service.PermissionService
	logger            *zap.Logger

	client *gotgproto.Client
}

func New(app *app.App) (*Bot, error) {
	var dcList dcs.List

	if app.Settings.Debug {
		dcList = dcs.Test()
	} else {
		dcList = dcs.Prod()
	}

	client, err := gotgproto.NewClient(
		app.Settings.AppId,
		app.Settings.AppHash,
		gotgproto.ClientTypeBot(app.Settings.BotToken),
		&gotgproto.ClientOpts{
			DCList:           dcList,
			DisableCopyright: true,
			InMemory:         true,
			Session:          sessionMaker.SimpleSession(),
			// Session:          sessionMaker.SqlSession(sqlite.Open("cache/muse-bot.db")),
		},
	)
	if err != nil {
		return nil, err
	}

	self := client.Self
	app.Logger.Info("bot logged in as https://t.me/" + self.Username)

	return &Bot{
		playlistService:   service.NewPlaylistService(app),
		permissionService: service.NewPermissionService(app),
		client:            client,
		logger:            app.Logger,
	}, nil
}

func (b *Bot) Setup() {
	disp := b.client.Dispatcher

	disp.AddHandler(handlers.NewChatMemberUpdated(nil, b.handleGroup))

	disp.AddHandler(handlers.NewCommand("start", b.handleStart))

	// handle group title and photo update
}

func (b *Bot) Start() error {
	return b.client.Idle()
}

func (b *Bot) Stop() {
	b.client.Stop()
}
