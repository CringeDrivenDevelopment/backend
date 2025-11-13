package dto

type Track struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Authors   string `json:"authors"`
	Thumbnail string `json:"thumbnail"`
	Length    int32  `json:"length"`
	Explicit  bool   `json:"explicit"`
}

type TrackAction struct {
	PlaylistId string `path:"playlist_id" minLength:"26" maxLength:"26" example:"01JZ35PYGP6HJA08H0NHYPBHWD" doc:"playlist id"`
	TrackId    string `path:"track_id" minLength:"11" maxLength:"11" example:"dQw4w9WgXcQ" doc:"track id"`
}

type SearchResponse struct {
	Body []Track
}
