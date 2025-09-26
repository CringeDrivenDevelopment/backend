package bot

import (
	"backend/internal/domain/dto"
	"backend/internal/domain/utils"
	"context"
	"errors"
	"slices"

	"github.com/celestix/gotgproto/functions"
	"github.com/gotd/td/tg"
)

func (b *Bot) iterateParticipants(ctx context.Context, channel *tg.InputChannel) (*[]utils.ParticipantData, error) {
	const limit = 100
	offset := 0
	var data []utils.ParticipantData

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
			var role string

			switch p := participant.(type) {
			case *tg.ChannelParticipant:
				userId = p.UserID
				role = dto.ViewerRole
			case *tg.ChannelParticipantCreator:
				userId = p.UserID
				role = dto.OwnerRole
			case *tg.ChannelParticipantAdmin:
				userId = p.UserID
				role = dto.ModeratorRole
			default:
				continue
			}

			if userId == b.client.Self.ID {
				continue
			}

			data = append(data, utils.ParticipantData{
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
			var role string

			switch p := participant.(type) {
			case *tg.ChannelParticipant:
				userId = p.UserID
				role = dto.ViewerRole
			case *tg.ChannelParticipantCreator:
				userId = p.UserID
				role = dto.OwnerRole
			case *tg.ChannelParticipantAdmin:
				userId = p.UserID
				role = dto.ModeratorRole
			default:
				continue
			}

			if userId == b.client.Self.ID {
				continue
			}

			if slices.Contains(data, utils.ParticipantData{
				UserID:  userId,
				NewRole: role,
				ChatID:  channel.ChannelID,
			}) {
				continue
			}

			data = append(data, utils.ParticipantData{
				UserID:  userId,
				NewRole: role,
				ChatID:  channel.ChannelID,
			})
		}

		offset += len(val.Participants)
	}

	return &data, nil
}

func (b *Bot) getChatInfo(ctx context.Context, chatID, actorID int64) (*utils.Chat, error) {
	peer, err := functions.GetInputPeerClassFromId(b.client.PeerStorage, chatID), error(nil)
	if peer == nil {
		return nil, errors.New("peer is nil")
	}
	if err != nil {
		return nil, err
	}

	var chat tg.ChatClass
	var info *utils.Chat
	var users *[]utils.ParticipantData

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
		var usersTemp []utils.ParticipantData
		for _, user := range chatFull.Users {
			if val, ok := user.(*tg.User); ok {
				if val.Bot {
					continue
				}

				if val.ID != actorID {
					usersTemp = append(usersTemp, utils.ParticipantData{
						NewRole: dto.ViewerRole,
						ChatID:  chatID,
						UserID:  val.ID,
					})
				} else {
					usersTemp = append(usersTemp, utils.ParticipantData{
						NewRole: dto.OwnerRole,
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
		info = &utils.Chat{
			Title: c.Title,
			Photo: c.Photo,
		}
		if users == nil || len(*users) == 0 {
			return info, errors.New("no chat users found")
		}
	case *tg.Channel:
		info = &utils.Chat{
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
