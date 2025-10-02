package models

import (
	"backend/internal/domain/queries"

	"github.com/gotd/td/tg"
)

type ParticipantData struct {
	PrevRole queries.PlaylistRole
	NewRole  queries.PlaylistRole
	UserID   int64
	ChatID   int64
	ActorID  int64
}

type Chat struct {
	Title string
	Photo tg.ChatPhotoClass
	Users *[]ParticipantData
}
