package bot

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func (b *Bot) handleGroup(context *ext.Context, update *ext.Update) error {
	val, ok := update.UpdateClass.(*tg.UpdateChannelParticipant)
	if !ok {
		return nil
	}

	prev := ""
	switch val.PrevParticipant.(type) {
	case *tg.ChannelParticipant:
		prev = "member"
	case *tg.ChannelParticipantSelf:
		prev = "self"
	case *tg.ChannelParticipantCreator:
		prev = "creator"
	case *tg.ChannelParticipantAdmin:
		prev = "admin"
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
		_, err := context.SendMessage(val.ChannelID, &tg.MessagesSendMessageRequest{Message: "дай админку пж"})
		return err
	}

	if prev == "self" && curr == "admin" {
		_, err := context.SendMessage(val.ChannelID, &tg.MessagesSendMessageRequest{Message: "респект за админку"})
		if err != nil {
			return err
		}

	}

	// todo: if self -> left = delete group
	// todo: if left -> self/admin = add group

	return nil
}
