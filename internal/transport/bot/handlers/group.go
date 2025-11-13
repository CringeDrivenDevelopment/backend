package handlers

import (
	"backend/internal/db/queries"
	"backend/internal/transport/bot/utils"
	"errors"
	"strconv"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/jackc/pgx/v5"
)

func (b *Bot) handleGroup(ctx *ext.Context, update *ext.Update) error {
	data, err := utils.HandleParticipant(update)
	if err != nil {
		b.logger.Error(err.Error())
		return err
	}

	b.logger.Info("Handling group update. ChatID: " + strconv.FormatInt(data.ChatID, 10) + ", UserID: " + strconv.FormatInt(data.UserID, 10) + ", PrevRole: " + string(data.PrevRole) + ", NewRole: " + string(data.NewRole))

	if data.UserID == b.client.Self.ID {
		if data.PrevRole == "" {
			_, err := ctx.SendMessage(data.ChatID, &tg.MessagesSendMessageRequest{Message: "Привет, я Лотти - бот для управления плейлистами!" + "\n\n" +
				"Можешь дать мне права администратора (чтобы я мог видеть лог AKA recent actions/недавние действия, а также всех участников и администраторов)"})
			if err != nil {
				b.logger.Error(err.Error())
			}
			return nil
		}

		if data.PrevRole == queries.PlaylistRoleViewer && data.NewRole == queries.PlaylistRoleModerator {
			_, err := ctx.SendMessage(data.ChatID, &tg.MessagesSendMessageRequest{Message: "Респект тебе за админку, сейчас создам плейлист!"})
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}

			// get basic chat info
			chat, err := utils.GetChatInfo(b.client, ctx.Context, data.ChatID, data.ActorID)
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}

			// TODO: handle group avatar, set type

			// create playlist
			create, err := b.playlistService.Create(ctx.Context, chat.Title, queries.PlaylistTypeYoutube, data.ChatID)
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}

			err = b.permissionService.AddGroup(ctx.Context, create.Id, *chat.Users)
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}

			return nil
		}

		if data.NewRole == "" || data.NewRole == queries.PlaylistRoleViewer {
			playlist, err := b.playlistService.GetByGroup(ctx, data.ChatID)
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}

			err = b.playlistService.Delete(ctx.Context, playlist.Id)
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}

			return nil
		}
	} else {
		playlist, err := b.playlistService.GetByGroup(ctx, data.ChatID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			b.logger.Error(err.Error())
			return err
		}

		if data.NewRole == "" {
			err = b.permissionService.Remove(ctx, playlist.Id, data.UserID)
			if err != nil {
				b.logger.Error(err.Error())
				return err
			}
		} else {
			if data.PrevRole == "" {
				err = b.permissionService.Add(ctx, data.NewRole, playlist.Id, data.UserID)
				if err != nil {
					b.logger.Error(err.Error())
					return err
				}
			} else {
				err = b.permissionService.Edit(ctx, data.NewRole, playlist.Id, data.UserID)
				if err != nil {
					b.logger.Error(err.Error())
					return err
				}
			}
		}
	}

	return nil
}
