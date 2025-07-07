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

	println("prev: ")
	switch val.PrevParticipant.(type) {
	case *tg.ChannelParticipant:
		println("member")
	case *tg.ChannelParticipantSelf:
		println("self")
	case *tg.ChannelParticipantCreator:
		println("creator")
	case *tg.ChannelParticipantAdmin:
		println("admin")
	default:
		println("left")
	}

	println("new: ")
	switch val.NewParticipant.(type) {
	case *tg.ChannelParticipant:
		println("member")
	case *tg.ChannelParticipantSelf:
		println("self")
	case *tg.ChannelParticipantCreator:
		println("creator")
	case *tg.ChannelParticipantAdmin: // channelParticipantAdmin#34c3bb53
		println("admin")
	default:
		println("left")
	}

	// todo: if self -> left = delete group
	// todo: if left -> self/admin = add group

	return nil
}
