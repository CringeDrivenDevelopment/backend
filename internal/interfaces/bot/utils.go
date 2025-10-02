package bot

import (
	"backend/internal/domain/models"
	"backend/internal/domain/queries"
	"context"
	"errors"
	"slices"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/functions"
	"github.com/gotd/td/tg"
)

func HandleParticipant(update *ext.Update) (models.ParticipantData, error) {
	var data models.ParticipantData
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

func extractChannelData(update *tg.UpdateChannelParticipant) models.ParticipantData {
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

func extractChatData(update *tg.UpdateChatParticipant) models.ParticipantData {
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

func (b *Bot) iterateParticipants(ctx context.Context, channel *tg.InputChannel) (*[]models.ParticipantData, error) {
	const limit = 100
	offset := 0
	var data []models.ParticipantData

	for offset%100 == 0 {
		resp, err := b.client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: channel,
			Filter:  &tg.ChannelParticipantsRecent{},
			Offset:  offset,
			Limit:   limit,
			Hash:    0,
		})
		if err != nil {
			return nil, err
		}

		val, ok := resp.(*tg.ChannelsChannelParticipants)
		if !ok {
			return nil, errors.New("invalid response " + resp.TypeName())
		}

		for _, participant := range val.Participants {
			var userId int64
			var role queries.PlaylistRole

			switch p := participant.(type) {
			case *tg.ChannelParticipant:
				userId = p.UserID
				role = queries.PlaylistRoleViewer
			case *tg.ChannelParticipantCreator:
				userId = p.UserID
				role = queries.PlaylistRoleOwner
			case *tg.ChannelParticipantAdmin:
				userId = p.UserID
				role = queries.PlaylistRoleModerator
			default:
				continue
			}

			if userId == b.client.Self.ID {
				continue
			}

			data = append(data, models.ParticipantData{
				UserID:  userId,
				NewRole: role,
				ChatID:  channel.ChannelID,
			})
		}

		offset += len(val.Participants)
	}

	offset = 0
	for offset%100 == 0 {
		resp, err := b.client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: channel,
			Filter:  &tg.ChannelParticipantsAdmins{},
			Offset:  offset,
			Limit:   limit,
			Hash:    0,
		})
		if err != nil {
			return nil, err
		}

		val, ok := resp.(*tg.ChannelsChannelParticipants)
		if !ok {
			return nil, errors.New("invalid response " + resp.TypeName())
		}

		for _, participant := range val.Participants {
			var userId int64
			var role queries.PlaylistRole

			switch p := participant.(type) {
			case *tg.ChannelParticipant:
				userId = p.UserID
				role = queries.PlaylistRoleViewer
			case *tg.ChannelParticipantCreator:
				userId = p.UserID
				role = queries.PlaylistRoleOwner
			case *tg.ChannelParticipantAdmin:
				userId = p.UserID
				role = queries.PlaylistRoleModerator
			default:
				continue
			}

			if userId == b.client.Self.ID {
				continue
			}

			if slices.Contains(data, models.ParticipantData{
				UserID:  userId,
				NewRole: role,
				ChatID:  channel.ChannelID,
			}) {
				continue
			}

			data = append(data, models.ParticipantData{
				UserID:  userId,
				NewRole: role,
				ChatID:  channel.ChannelID,
			})
		}

		offset += len(val.Participants)
	}

	return &data, nil
}

func (b *Bot) getChatInfo(ctx context.Context, chatID, actorID int64) (*models.Chat, error) {
	peer, err := functions.GetInputPeerClassFromId(b.client.PeerStorage, chatID), error(nil)
	if peer == nil {
		return nil, errors.New("peer is nil")
	}
	if err != nil {
		return nil, err
	}

	var chat tg.ChatClass
	var info *models.Chat
	var users *[]models.ParticipantData

	switch peerResult := peer.(type) {
	case *tg.InputPeerChat:
		chatFull, err := b.client.API().MessagesGetFullChat(ctx, chatID)
		if err != nil {
			return nil, err
		}

		if len(chatFull.Chats) != 1 {
			return nil, errors.New("no chat found")
		}

		chat = chatFull.Chats[0]
		var usersTemp []models.ParticipantData
		for _, user := range chatFull.Users {
			if val, ok := user.(*tg.User); ok {
				if val.Bot {
					continue
				}

				if val.ID != actorID {
					usersTemp = append(usersTemp, models.ParticipantData{
						NewRole: queries.PlaylistRoleViewer,
						ChatID:  chatID,
						UserID:  val.ID,
					})
				} else {
					usersTemp = append(usersTemp, models.ParticipantData{
						NewRole: queries.PlaylistRoleOwner,
						ChatID:  chatID,
						UserID:  val.ID,
					})
				}
			}
		}
		users = &usersTemp
	case *tg.InputPeerChannel:
		channelFull, err := b.client.API().ChannelsGetFullChannel(ctx, &tg.InputChannel{
			ChannelID:  peerResult.ChannelID,
			AccessHash: peerResult.AccessHash,
		})
		if err != nil {
			return nil, err
		}

		if len(channelFull.Chats) != 1 {
			return nil, errors.New("no chat found")
		}

		chat = channelFull.Chats[0]
	default:
		return nil, errors.New("unknown peer type " + peerResult.TypeName())
	}

	if chat == nil {
		return nil, errors.New("no chat found")
	}

	switch c := chat.(type) {
	case *tg.Chat:
		info = &models.Chat{
			Title: c.Title,
			Photo: c.Photo,
		}
		if users == nil || len(*users) == 0 {
			return info, errors.New("no chat users found")
		}
	case *tg.Channel:
		info = &models.Chat{
			Title: c.Title,
			Photo: c.Photo,
		}

		users, err = b.iterateParticipants(ctx, &tg.InputChannel{
			AccessHash: c.AccessHash,
			ChannelID:  c.ID,
		})

		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown channel type " + c.TypeName())
	}

	info.Users = users

	return info, nil
}
