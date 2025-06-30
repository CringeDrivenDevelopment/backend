package dto

type Playlist struct {
	Id            string
	Title         string
	Thumbnail     string
	Length        int
	Tracks        *[]Track
	AllowedLength int
	AllowedTracks *[]Track
}
