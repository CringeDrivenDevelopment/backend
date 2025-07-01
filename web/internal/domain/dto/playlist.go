package dto

type Playlist struct {
	Id            string   `json:"id"`
	Title         string   `json:"title"`
	Thumbnail     string   `json:"thumbnail"`
	Length        int      `json:"length"`
	Tracks        *[]Track `json:"tracks"`
	AllowedLength int      `json:"allowed_length"`
	AllowedTracks *[]Track `json:"allowed_tracks"`
}
