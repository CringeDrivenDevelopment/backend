package dto

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

var UserRoles = []string{ViewerRole, ModeratorRole, OwnerRole}
