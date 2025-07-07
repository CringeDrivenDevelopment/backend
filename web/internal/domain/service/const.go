package service

/*
CREATE TYPE playlist_role AS ENUM ('viewer', 'moderator', 'owner');
*/
const (
	GroupRole     = "group"
	ViewerRole    = "viewer"
	ModeratorRole = "moderator"
	OwnerRole     = "owner"

	CustomSource = "custom"
	TgSource     = "tg"
)

var roles = []string{ViewerRole, ModeratorRole, OwnerRole}
