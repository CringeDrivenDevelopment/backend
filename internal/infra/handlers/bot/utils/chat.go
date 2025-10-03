package utils

import (
	"backend/internal/application/models"
	"backend/internal/infra/database/queries"
	"context"
	"errors"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/functions"
	"github.com/gotd/td/tg"
)

func GetChatInfo(client *gotgproto.Client, ctx context.Context, chatID, actorID int64) (*models.Chat, error) {
	peer, err := functions.GetInputPeerClassFromId(client.PeerStorage, chatID), error(nil)
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
		chatFull, err := client.API().MessagesGetFullChat(ctx, chatID)
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
		channelFull, err := client.API().ChannelsGetFullChannel(ctx, &tg.InputChannel{
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

		users, err = iterateParticipants(client, ctx, &tg.InputChannel{
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
