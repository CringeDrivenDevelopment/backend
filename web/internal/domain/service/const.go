package service

/*
CREATE TYPE playlist_role AS ENUM ('viewer', 'moderator', 'owner');
*/
const ViewerRole = "viewer"
const ModeratorRole = "moderator"
const OwnerRole = "owner"

var roles = []string{ViewerRole, ModeratorRole, OwnerRole}

const CustomSource = "custom"
const TgSource = "tg"
