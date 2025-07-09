package dto

type Playlist struct {
	Id            string   `json:"id"`
	Title         string   `json:"title"`
	Thumbnail     string   `json:"thumbnail"`
	Tracks        []Track  `json:"tracks,omitempty"`
	AllowedIds    []string `json:"allowed_ids,omitempty"`
	Count         int      `json:"count"`
	AllowedCount  int      `json:"allowed_count"`
	Length        int      `json:"length"`
	AllowedLength int      `json:"allowed_length"`
	Role          string   `json:"role"`
	Type          string   `json:"type"`
}

type Archive struct {
	Filename string `json:"filename"`
}
