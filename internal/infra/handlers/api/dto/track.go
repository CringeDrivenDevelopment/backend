package dto

type Track struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Authors     string   `json:"authors"`
	Thumbnail   string   `json:"thumbnail"`
	Length      int32    `json:"length"`
	Explicit    bool     `json:"explicit"`
	PlaylistIds []string `json:"playlist_ids,omitempty"`
}
