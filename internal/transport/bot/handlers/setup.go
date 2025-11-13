package handlers

import (
	"backend/internal/infra"
	"backend/internal/interfaces"
	"backend/internal/service"

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
	playlistService   interfaces.PlaylistService
	permissionService interfaces.PermissionService
	logger            *zap.Logger
	// dl                *downloader.Downloader
	// s3                *service.S3Service

	client *gotgproto.Client
}

func New(cfg *infra.Config, playlistService *service.Playlist, permissionService *service.Permission, logger *zap.Logger) (*Bot, error) {
	var dcList dcs.List

	if cfg.Debug {
		dcList = dcs.Test()
	} else {
		dcList = dcs.Prod()
	}

	client, err := gotgproto.NewClient(
		cfg.AppId,
		cfg.AppHash,
		gotgproto.ClientTypeBot(cfg.BotToken),
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
	logger.Info("bot logged in as https://t.me/" + self.Username)

	/*
		s3Service, err := service.NewS3Service(app)
		if err != nil {
			return nil, err
		}
	*/

	return &Bot{
		playlistService:   playlistService,
		permissionService: permissionService,
		// 	dl:                downloader.NewDownloader(),
		//	s3:                s3Service,
		client: client,
		logger: logger,
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
