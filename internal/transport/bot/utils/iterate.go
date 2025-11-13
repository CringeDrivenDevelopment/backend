package utils

import (
	"backend/internal/db/queries"
	"backend/internal/transport/bot/models"
	"context"
	"errors"
	"slices"

	"github.com/celestix/gotgproto"
	"github.com/gotd/td/tg"
)

func iterateParticipants(client *gotgproto.Client, ctx context.Context, channel *tg.InputChannel) (*[]models.ParticipantData, error) {
	const limit = 100
	offset := 0
	var data []models.ParticipantData

	for offset%100 == 0 {
		resp, err := client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
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

			if userId == client.Self.ID {
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
		resp, err := client.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
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

			if userId == client.Self.ID {
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
