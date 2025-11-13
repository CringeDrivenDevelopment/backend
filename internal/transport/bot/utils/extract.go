package utils

import (
	"backend/internal/db/queries"
	"backend/internal/transport/bot/models"

	"github.com/gotd/td/tg"
)

func ExtractChannelData(update *tg.UpdateChannelParticipant) models.ParticipantData {
	var newRole queries.PlaylistRole
	var prevRole queries.PlaylistRole

	if update.PrevParticipant != nil {
		switch update.PrevParticipant.(type) {
		case *tg.ChannelParticipant:
			prevRole = queries.PlaylistRoleViewer
		case *tg.ChannelParticipantAdmin:
			prevRole = queries.PlaylistRoleModerator
		case *tg.ChannelParticipantCreator:
			prevRole = queries.PlaylistRoleOwner
		case *tg.ChannelParticipantSelf:
			prevRole = queries.PlaylistRoleViewer
		default:
			prevRole = ""
		}
	}

	if update.NewParticipant != nil {
		switch update.NewParticipant.(type) {
		case *tg.ChannelParticipant:
			newRole = queries.PlaylistRoleViewer
		case *tg.ChannelParticipantAdmin:
			newRole = queries.PlaylistRoleModerator
		case *tg.ChannelParticipantCreator:
			newRole = queries.PlaylistRoleOwner
		case *tg.ChannelParticipantSelf:
			newRole = queries.PlaylistRoleViewer
		default:
			newRole = ""
		}
	}

	return models.ParticipantData{
		PrevRole: prevRole,
		NewRole:  newRole,
		UserID:   update.UserID,
		ChatID:   update.ChannelID,
		ActorID:  update.ActorID,
	}
}

func ExtractChatData(update *tg.UpdateChatParticipant) models.ParticipantData {
	var newRole queries.PlaylistRole
	var prevRole queries.PlaylistRole

	if update.PrevParticipant != nil {
		switch update.PrevParticipant.(type) {
		case *tg.ChatParticipant:
			prevRole = queries.PlaylistRoleViewer
		case *tg.ChatParticipantAdmin:
			prevRole = queries.PlaylistRoleModerator
		case *tg.ChatParticipantCreator:
			prevRole = queries.PlaylistRoleOwner
		default:
			prevRole = ""
		}
	}

	if update.NewParticipant != nil {
		switch update.NewParticipant.(type) {
		case *tg.ChatParticipant:
			newRole = queries.PlaylistRoleViewer
		case *tg.ChatParticipantAdmin:
			newRole = queries.PlaylistRoleModerator
		case *tg.ChatParticipantCreator:
			newRole = queries.PlaylistRoleOwner
		default:
			newRole = ""
		}
	}

	return models.ParticipantData{
		PrevRole: prevRole,
		NewRole:  newRole,
		UserID:   update.UserID,
		ChatID:   update.ChatID,
		ActorID:  update.ActorID,
	}
}
