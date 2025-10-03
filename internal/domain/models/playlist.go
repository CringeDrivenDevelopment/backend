package models

import (
	"backend/internal/infra/database/queries"
)

type DtoPlaylist struct {
	Id            string               `json:"id"`
	Title         string               `json:"title"`
	Thumbnail     string               `json:"thumbnail"`
	Tracks        []DtoTrack           `json:"tracks,omitempty"`
	AllowedIds    []string             `json:"allowed_ids,omitempty"`
	Count         int                  `json:"count"`
	AllowedCount  int                  `json:"allowed_count"`
	Length        int                  `json:"length"`
	AllowedLength int                  `json:"allowed_length"`
	Role          queries.PlaylistRole `json:"role"`
	Type          string               `json:"type"`
}
