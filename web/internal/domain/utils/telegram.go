package utils

import (
	"backend/internal/domain/dto"
	"errors"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

type ParticipantData struct {
	PrevRole string
	NewRole  string
	UserID   int64
	ChatID   int64
	ActorID  int64
}

type Chat struct {
	Title string
	Photo tg.ChatPhotoClass
	Users *[]ParticipantData
}

func HandleParticipant(update *ext.Update) (ParticipantData, error) {
	var data ParticipantData
	switch u := update.UpdateClass.(type) {
	case *tg.UpdateChannelParticipant:
		data = extractChannelData(u)
	case *tg.UpdateChatParticipant:
		data = extractChatData(u)
	default:
		return data, errors.New("invalid update type " + u.TypeName())
	}

	return data, nil
}

func extractChannelData(update *tg.UpdateChannelParticipant) ParticipantData {
	newRole := ""
	prevRole := ""

	if update.PrevParticipant != nil {
		switch update.PrevParticipant.(type) {
		case *tg.ChannelParticipant:
			prevRole = dto.ViewerRole
		case *tg.ChannelParticipantAdmin:
			prevRole = dto.ModeratorRole
		case *tg.ChannelParticipantCreator:
			prevRole = dto.OwnerRole
		case *tg.ChannelParticipantSelf:
			prevRole = dto.ViewerRole
		default:
			prevRole = ""
		}
	}

	if update.NewParticipant != nil {
		switch update.NewParticipant.(type) {
		case *tg.ChannelParticipant:
			newRole = dto.ViewerRole
		case *tg.ChannelParticipantAdmin:
			newRole = dto.ModeratorRole
		case *tg.ChannelParticipantCreator:
			newRole = dto.OwnerRole
		case *tg.ChannelParticipantSelf:
			newRole = dto.ViewerRole
		default:
			newRole = ""
		}
	}

	return ParticipantData{
		PrevRole: prevRole,
		NewRole:  newRole,
		UserID:   update.UserID,
		ChatID:   update.ChannelID,
		ActorID:  update.ActorID,
	}
}

func extractChatData(update *tg.UpdateChatParticipant) ParticipantData {
	prevRole := ""
	newRole := ""

	if update.PrevParticipant != nil {
		switch update.PrevParticipant.(type) {
		case *tg.ChatParticipant:
			prevRole = dto.ViewerRole
		case *tg.ChatParticipantAdmin:
			prevRole = dto.ModeratorRole
		case *tg.ChatParticipantCreator:
			prevRole = dto.OwnerRole
		default:
			prevRole = ""
		}
	}

	if update.NewParticipant != nil {
		switch update.NewParticipant.(type) {
		case *tg.ChatParticipant:
			newRole = dto.ViewerRole
		case *tg.ChatParticipantAdmin:
			newRole = dto.ModeratorRole
		case *tg.ChatParticipantCreator:
			newRole = dto.OwnerRole
		default:
			newRole = ""
		}
	}

	return ParticipantData{
		PrevRole: prevRole,
		NewRole:  newRole,
		UserID:   update.UserID,
		ChatID:   update.ChatID,
		ActorID:  update.ActorID,
	}
}
