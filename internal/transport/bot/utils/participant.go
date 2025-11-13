package utils

import (
	"backend/internal/transport/bot/models"
	"errors"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
)

func HandleParticipant(update *ext.Update) (models.ParticipantData, error) {
	var data models.ParticipantData
	switch u := update.UpdateClass.(type) {
	case *tg.UpdateChannelParticipant:
		data = ExtractChannelData(u)
	case *tg.UpdateChatParticipant:
		data = ExtractChatData(u)
	default:
		return data, errors.New("invalid update type " + u.TypeName())
	}

	return data, nil
}
