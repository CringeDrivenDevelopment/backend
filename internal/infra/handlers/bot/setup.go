package bot

import (
	"backend/internal/application"
	"backend/internal/application/service"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
)

type Bot struct {
	playlistService   *service.Playlist
	permissionService *service.Permission
	logger            *zap.Logger
	// dl                *downloader.Downloader
	// s3                *service.S3Service

	client *gotgproto.Client
}

func New(app *application.App) (*Bot, error) {
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
			Session:          sessionMaker.SqlSession(sqlite.Open("telegram/bot.db")),
		},
	)
	if err != nil {
		return nil, err
	}

	self := client.Self
	app.Logger.Info("bot logged in as https://t.me/" + self.Username)

	/*
		s3Service, err := service.NewS3Service(app)
		if err != nil {
			return nil, err
		}
	*/

	return &Bot{
		playlistService:   service.NewPlaylistService(app),
		permissionService: service.NewPermissionService(app),
		// 	dl:                downloader.NewDownloader(),
		//	s3:                s3Service,
		client: client,
		logger: app.Logger,
	}, nil
}

func (b *Bot) Setup() {
	disp := b.client.Dispatcher

	disp.AddHandler(handlers.NewChatMemberUpdated(nil, b.handleGroup))

	disp.AddHandler(handlers.NewCommand("start", b.handleStart))

	disp.AddHandler(handlers.NewMessage(func(msg *types.Message) bool {
		_, okTitle := msg.Action.(*tg.MessageActionChatEditTitle)
		// _, okPhoto := msg.Action.(*tg.MessageActionChatEditPhoto)
		return okTitle // || okPhoto
	}, b.handleChatAction))
}

func (b *Bot) Start() error {
	return b.client.Idle()
}

func (b *Bot) Stop() {
	b.client.Stop()
}
