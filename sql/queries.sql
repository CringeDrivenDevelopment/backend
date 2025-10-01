-- name: CreatePlaylist :exec
INSERT INTO playlists (id, title, thumbnail, tracks, allowed_tracks, type, external_id)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: EditPlaylist :exec
UPDATE playlists
SET
    title = COALESCE($2, title),
    thumbnail = COALESCE($3, thumbnail),
    tracks = COALESCE($4, tracks),
    allowed_tracks = COALESCE($5, allowed_tracks),
    type = COALESCE($6, type),
    external_id = COALESCE($7, external_id)
WHERE id = $1;

-- name: DeletePlaylist :exec
DELETE FROM playlists WHERE id = $1;

-- name: GetUserPlaylists :many
SELECT
    pl.*,
    p.role
FROM playlists pl
         JOIN playlist_permissions p ON pl.id = p.playlist_id
         JOIN users u ON p.user_id = u.id  -- Join users table
WHERE p.user_id = $1;

-- name: GetPlaylistById :one
SELECT
    pl.*,
    p.role
FROM playlist_permissions p
         JOIN playlists pl ON p.playlist_id = pl.id
         JOIN users u ON p.user_id = u.id  -- Join users table
WHERE p.playlist_id = $1 AND  p.user_id = $2;

-- name: GetTrackPlaylists :many
-- param: TrackId text
-- param: UserId bigint
SELECT pl.id
FROM playlists pl
         JOIN playlist_permissions pp ON pl.id = pp.playlist_id
WHERE
    pp.user_id = sqlc.arg(user_id)
  AND sqlc.arg(track_id)::text = ANY(pl.tracks);

-- name: CreateUser :exec
INSERT INTO users (id) VALUES ($1);

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: CreateTrack :exec
INSERT INTO tracks (id, title, authors, thumbnail, length, explicit)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTrackById :one
SELECT * FROM tracks WHERE id = $1;

-- name: CreateRole :exec
INSERT INTO playlist_permissions (playlist_id, user_id, role)
VALUES ($1, $2, $3);

-- name: EditRole :exec
UPDATE playlist_permissions
SET role = $3
WHERE playlist_id = $1 AND user_id = $2;

-- name: DeleteRole :exec
DELETE FROM playlist_permissions
WHERE playlist_id = $1 AND user_id = $2;

-- name: GetRole :one
SELECT playlist_id FROM playlist_permissions
WHERE user_id = $1 AND role = $2;

-- ROLES CRUD END
