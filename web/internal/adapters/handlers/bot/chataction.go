package bot

import (
	"backend/internal/domain/dto"
	"context"
	"errors"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/jackc/pgx/v5"
	"io"
	"strconv"
)

func (b *Bot) handleChatAction(ctx *ext.Context, update *ext.Update) error {
	var msg tg.MessageClass
	switch um := update.UpdateClass.(type) {
	case *tg.UpdateNewMessage:
		msg = um.Message
	case *tg.UpdateNewChannelMessage:
		msg = um.Message
	default:
		return nil
	}

	serviceMessage, ok := msg.(*tg.MessageService)
	if !ok {
		return nil
	}

	var id int64
	switch peer := serviceMessage.PeerID.(type) {
	case *tg.PeerChat:
		id = peer.ChatID
	case *tg.PeerChannel:
		id = peer.ChannelID
	default:
		return nil
	}

	b.logger.Info("handle service message: " + serviceMessage.TypeName() + ", id: " + strconv.Itoa(serviceMessage.ID) + ", chatID: " + strconv.FormatInt(id, 10))

	switch smResult := serviceMessage.Action.(type) {
	case *tg.MessageActionChatEditTitle:
		err := b.handleTitleUpdate(ctx.Context, smResult.Title, id)
		if err != nil {
			b.logger.Error(err.Error())
		}
	case *tg.MessageActionChatEditPhoto:
		err := b.handlePhotoUpdate(ctx, smResult.Photo, id)
		if err != nil {
			b.logger.Error(err.Error())
		}
	}

	return nil
}

func (b *Bot) handlePhotoUpdate(ctx context.Context, photo tg.PhotoClass, chatID int64) error {
	playlistID, err := b.permissionService.Get(ctx, chatID, dto.GroupRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		b.logger.Error(err.Error())
		return err
	}

	switch photoResult := photo.(type) {
	case *tg.Photo:
		bestSize := photoResult.Sizes[len(photoResult.Sizes)-1].(*tg.PhotoSize)

		if bestSize == nil {
			return errors.New("no suitable size found")
		}

		loc := &tg.InputPhotoFileLocation{
			ID:            photoResult.ID,
			AccessHash:    photoResult.AccessHash,
			FileReference: photoResult.FileReference,
			ThumbSize:     bestSize.Type,
		}

		pr, pw := io.Pipe()
		_, err = b.dl.Download(b.client.Client.API(), loc).Stream(ctx, pw)
		if err != nil {
			return err
		}
		err = pw.Close()
		if err != nil {
			return err
		}

		_, err = b.s3.UploadPhoto(ctx, playlistID, pr)
		if err != nil {
			return err
		}

		err = pr.Close()
		if err != nil {
			return err
		}

		err = b.playlistService.UpdatePhoto(ctx, playlistID, playlistID, chatID)
		if err != nil {
			return err
		}
	case *tg.PhotoEmpty:
		err = b.s3.DeletePhoto(ctx, playlistID)
		if err != nil {
			return err
		}

		err = b.playlistService.UpdatePhoto(ctx, playlistID, "", chatID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bot) handleTitleUpdate(ctx context.Context, title string, chatID int64) error {
	playlistID, err := b.permissionService.Get(ctx, chatID, dto.GroupRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		b.logger.Error(err.Error())
		return err
	}

	err = b.playlistService.Rename(ctx, playlistID, title, chatID)
	if err != nil {
		return err
	}

	return nil
}
