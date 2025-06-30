-- INIT DATABASE TABLES / SCHEMA BEGIN

-- name: InitPlaylists :exec
CREATE TABLE IF NOT EXISTS playlists (
    id TEXT NOT NULL PRIMARY KEY UNIQUE,
    title TEXT NOT NULL,
    thumbnail TEXT NOT NULL,
    tracks TEXT[] DEFAULT '{}',
    allowed_tracks TEXT[] DEFAULT '{}',
    length INTEGER GENERATED ALWAYS AS (COALESCE(array_length(tracks, 1), 0)) STORED,
    allowed_length INTEGER GENERATED ALWAYS AS (COALESCE(array_length(allowed_tracks, 1), 0)) STORED
);

-- name: InitTracks :exec
CREATE TABLE IF NOT EXISTS tracks (
    id TEXT NOT NULL PRIMARY KEY,
    title TEXT NOT NULL,
    authors TEXT NOT NULL,
    thumbnail TEXT NOT NULL,
    length INTEGER NOT NULL,
    explicit BOOLEAN NOT NULL DEFAULT FALSE
);

-- name: InitRoleEnum :exec
DO $$ BEGIN
    CREATE TYPE playlist_role AS ENUM ('viewer', 'moderator', 'owner');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- name: InitPermissions :exec
CREATE TABLE IF NOT EXISTS playlist_permissions (
    playlist_id TEXT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    role playlist_role NOT NULL,
    PRIMARY KEY (playlist_id, user_id)
);

-- name: InitUsers :exec
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL
);

-- INIT DATABASE TABLES / SCHEMA END


-- =================================


-- PLAYLISTS CRUD BEGIN

-- name: CreatePlaylist :exec
INSERT INTO playlists (id, title, thumbnail, tracks, allowed_tracks)
VALUES ($1, $2, $3, $4, $5);

-- name: EditPlaylist :exec
UPDATE playlists
SET
    title = COALESCE($2, title),
    thumbnail = COALESCE($3, thumbnail),
    tracks = COALESCE($4, tracks),
    allowed_tracks = COALESCE($5, allowed_tracks)
WHERE id = $1;

-- name: DeletePlaylist :exec
DELETE FROM playlists WHERE id = $1;

-- name: GetPlaylistById :one
SELECT * FROM playlists WHERE id = $1;

-- name: GetUserPlaylists :many
SELECT
    pl.*,
    p.role,
    u.name AS user_name  -- New: include user name
FROM playlists pl
         JOIN playlist_permissions p ON pl.id = p.playlist_id
         JOIN users u ON p.user_id = u.id  -- Join users table
WHERE p.user_id = $1;

-- name: GetPlaylistByUserId :many
SELECT
    p.playlist_id,
    p.user_id,
    u.name AS user_name,  -- New: include user name
    p.role,
    pl.title AS playlist_title,
    pl.thumbnail,
    pl.length AS track_count,
    pl.allowed_length AS allowed_count
FROM playlist_permissions p
         JOIN playlists pl ON p.playlist_id = pl.id
         JOIN users u ON p.user_id = u.id  -- Join users table
WHERE p.user_id = $1;

-- name: GetPlaylistByPlaylistId :many
SELECT
    p.playlist_id,
    p.user_id,
    u.name AS user_name,  -- New: include user name
    p.role,
    pl.title AS playlist_title,
    pl.thumbnail,
    pl.length AS track_count,
    pl.allowed_length AS allowed_count
FROM playlist_permissions p
         JOIN playlists pl ON p.playlist_id = pl.id
         JOIN users u ON p.user_id = u.id  -- Join users table
WHERE p.playlist_id = $1;

-- PLAYLISTS CRUD END


-- =================================


-- USERS CRUD BEGIN

-- name: CreateUser :exec
INSERT INTO users (id, name) VALUES ($1, $2);

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: EditUser :exec
UPDATE users SET name = $2 WHERE id = $1;

-- USERS CRUD END


-- =================================


-- TRACKS CRUD BEGIN

-- name: CreateTrack :exec
INSERT INTO tracks (id, title, authors, thumbnail, length, explicit)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetTrackById :one
SELECT * FROM tracks WHERE id = $1;

-- TRACKS CRUD END


-- =================================


-- ROLES CRUD BEGIN

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

-- ROLES CRUD END
