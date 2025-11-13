package dto

import (
	"backend/internal/infra/queries"
)

type Playlist struct {
	Id           string               `json:"id"`
	Title        string               `json:"title"`
	Thumbnail    string               `json:"thumbnail"`
	Tracks       []Track              `json:"tracks,omitempty"`
	AllowedIds   []string             `json:"allowed_ids,omitempty"`
	Count        int                  `json:"count"`
	AllowedCount int                  `json:"allowed_count"`
	Length       int                  `json:"length"`
	Role         queries.PlaylistRole `json:"role"`
	Type         string               `json:"type"`
}

type PlaylistByIdResponse struct {
	Body Playlist
}

type PlaylistsResponse struct {
	Body []Playlist
}
