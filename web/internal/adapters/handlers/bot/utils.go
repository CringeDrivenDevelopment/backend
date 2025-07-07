package bot

import (
	"backend/internal/domain/service"
	"context"
	"errors"
	"github.com/celestix/gotgproto/functions"
	"github.com/gotd/td/tg"
)

type chat struct {
	title string
	photo tg.ChatPhotoClass
}

func (b *Bot) getInputChannel(chatId int64) (*tg.InputChannel, error) {
	peer, err := functions.GetInputPeerClassFromId(b.client.PeerStorage, chatId), error(nil)
	if peer == nil {
		return nil, errors.New("peer is nil")
	}
	if err != nil {
		return nil, err
	}

	var inputChannel *tg.InputChannel
	switch p := peer.(type) {
	case *tg.InputPeerChannel:
		inputChannel = &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		}
	default:
		return nil, errors.New("invalid peer")
	}

	return inputChannel, nil
}

func (b *Bot) iterateParticipants(ctx context.Context, chanel *tg.InputChannel, playlistId string) error {
	const limit = 100
	offset := 0

	for {
		resp, err := b.client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: chanel,
			Filter:  &tg.ChannelParticipantsRecent{},
			Offset:  offset,
			Limit:   limit,
			Hash:    0,
		})
		if err != nil {
			return err
		}

		val, ok := resp.(*tg.ChannelsChannelParticipants)
		if !ok {
			return nil
		}

		for _, participant := range val.Participants {
			var userId int64
			var role string

			switch participant.(type) {
			case *tg.ChannelParticipant:
				userId = participant.(*tg.ChannelParticipant).UserID
				role = service.ViewerRole
			case *tg.ChannelParticipantCreator:
				userId = participant.(*tg.ChannelParticipantCreator).UserID
				role = service.OwnerRole
			case *tg.ChannelParticipantAdmin:
				userId = participant.(*tg.ChannelParticipantAdmin).UserID
				role = service.ModeratorRole
			default:
				continue
			}

			err := b.permissionService.Add(ctx, role, playlistId, userId)
			if err != nil {
				return err
			}
		}

		offset += limit
	}
}

func (b *Bot) getChatInfo(ctx context.Context, inputChannel *tg.InputChannel) (*chat, error) {
	full, err := b.client.API().ChannelsGetFullChannel(ctx, inputChannel)
	if err != nil {
		return nil, err
	}

	if len(full.Chats) == 0 {
		return nil, errors.New("no chat found")
	}

	if len(full.Chats) > 1 {
		return nil, errors.New("multiple chat found")
	}

	chatResult := full.Chats[0]
	switch chatResult.(type) {
	case *tg.Channel:
		result := chatResult.(*tg.Channel)
		return &chat{
			title: result.Title,
			photo: result.Photo,
		}, nil
	case *tg.Chat:
		result := chatResult.(*tg.Chat)
		return &chat{
			title: result.Title,
			photo: result.Photo,
		}, nil
	default:
		return nil, errors.New("invalid chat type " + chatResult.TypeName())
	}
}
