package bot

import (
	"backend/internal/domain/service"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func (b *Bot) handleGroup(ctx *ext.Context, update *ext.Update) error {
	val, ok := update.UpdateClass.(*tg.UpdateChannelParticipant)
	if !ok {
		return nil
	}

	prev := ""
	var id int64
	switch val.PrevParticipant.(type) {
	case *tg.ChannelParticipant:
		prev = "member"
		id = val.PrevParticipant.(*tg.ChannelParticipant).UserID
	case *tg.ChannelParticipantSelf:
		prev = "self"
		id = val.PrevParticipant.(*tg.ChannelParticipantSelf).UserID
	case *tg.ChannelParticipantCreator:
		prev = "creator"
		id = val.PrevParticipant.(*tg.ChannelParticipantCreator).UserID
	case *tg.ChannelParticipantAdmin:
		prev = "admin"
		id = val.PrevParticipant.(*tg.ChannelParticipantAdmin).UserID
	default:
		prev = "left"
	}

	curr := ""
	switch val.NewParticipant.(type) {
	case *tg.ChannelParticipant:
		curr = "member"
	case *tg.ChannelParticipantSelf:
		curr = "self"
	case *tg.ChannelParticipantCreator:
		curr = "creator"
	case *tg.ChannelParticipantAdmin: // channelParticipantAdmin#34c3bb53
		curr = "admin"
	default:
		curr = "left"
	}

	if prev == "left" && curr == "self" {
		_, err := ctx.SendMessage(val.ChannelID, &tg.MessagesSendMessageRequest{Message: "Привет, я Лотти - бот для управления плейлистами! \n" +
			"Можешь дать мне права администратора (чтобы я мог видеть лог AKA recent actions/недавние действия, а также всех участников и администраторов)"})
		return err
	}

	if prev == "self" && curr == "admin" {
		_, err := ctx.SendMessage(val.ChannelID, &tg.MessagesSendMessageRequest{Message: "Респект тебе за админку, сейчас создам плейлист!"})
		if err != nil {
			return err
		}

		// get input channel
		inputChat, err := b.getInputChannel(val.ChannelID)
		if err != nil {
			return err
		}

		// get basic chat info
		chat, err := b.getChatInfo(ctx.Context, inputChat)
		if err != nil {
			return err
		}

		// TODO: handle group avatar

		// create playlist
		create, err := b.playlistService.Create(ctx.Context, chat.title, service.TgSource)
		if err != nil {
			return err
		}

		// add roles to users, sorry(((((
		err = b.iterateParticipants(ctx.Context, inputChat, create.Id)
		if err != nil {
			return err
		}

		// add group to indexing
		err = b.permissionService.Add(ctx.Context, service.GroupRole, create.Id, val.ChannelID)
		if err != nil {
			return err
		}

		return nil
	}

	if curr == "left" && id == b.client.Self.ID {
		id, err := b.permissionService.Get(ctx, val.ChannelID, service.GroupRole)
		if err != nil {
			return err
		}

		err = b.playlistService.Delete(ctx.Context, id)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}
