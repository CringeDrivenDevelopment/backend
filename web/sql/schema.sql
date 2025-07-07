-- INIT DATABASE TABLES / SCHEMA BEGIN

-- name: InitPlaylists :exec
CREATE TABLE IF NOT EXISTS playlists (
    id TEXT NOT NULL PRIMARY KEY UNIQUE,
    title TEXT NOT NULL,
    thumbnail TEXT NOT NULL,
    tracks TEXT[] DEFAULT '{}',
    allowed_tracks TEXT[] DEFAULT '{}',
    count INTEGER GENERATED ALWAYS AS (COALESCE(array_length(tracks, 1), 0)) STORED,
    allowed_count INTEGER GENERATED ALWAYS AS (COALESCE(array_length(allowed_tracks, 1), 0)) STORED,
    type TEXT NOT NULL,
    time INTEGER NOT NULL DEFAULT 0,
    allowed_time INTEGER NOT NULL DEFAULT 0
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

-- name: InitPermissions :exec
CREATE TABLE IF NOT EXISTS playlist_permissions (
    playlist_id TEXT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    role TEXT NOT NULL,
    PRIMARY KEY (playlist_id, user_id)
);

-- name: InitUsers :exec
CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL
);

-- name: InitTracksIndex :exec
CREATE INDEX IF NOT EXISTS idx_tracks_id ON tracks (id);

-- name: InitPlaylistsTracksIndex :exec
CREATE INDEX IF NOT EXISTS idx_playlists_tracks ON playlists USING GIN(tracks);

-- name: InitPlaylistsAllowedTracksIndex :exec
CREATE INDEX IF NOT EXISTS idx_playlists_allowed_tracks ON playlists USING GIN(allowed_tracks);

-- name: InitPermissionsIndex :exec
CREATE INDEX IF NOT EXISTS idx_permissions_user ON playlist_permissions (user_id);

-- name: InitCalculatePlaylistTime :exec
CREATE OR REPLACE FUNCTION calculate_playlist_time(track_ids TEXT[])
    RETURNS INTEGER AS $$
DECLARE
    total_time INTEGER;
BEGIN
    SELECT COALESCE(SUM(length), 0) INTO total_time
    FROM tracks
    WHERE id = ANY(track_ids);

    RETURN total_time;
END;
$$ LANGUAGE plpgsql;

-- name: InitUpdatePlaylistTimes :exec
CREATE OR REPLACE FUNCTION update_playlist_times()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.time = calculate_playlist_time(NEW.tracks);

    NEW.allowed_time = calculate_playlist_time(NEW.allowed_tracks);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- name: InitPlaylistTimesTrigger :exec
DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1 FROM pg_trigger
            WHERE tgname = 'playlist_times_trigger'
        ) THEN
            CREATE TRIGGER playlist_times_trigger
                BEFORE INSERT OR UPDATE OF tracks, allowed_tracks ON playlists
                FOR EACH ROW
            EXECUTE FUNCTION update_playlist_times();
        END IF;
    END
$$;

-- name: InitUpdatePlaylistOnTrackChange :exec
CREATE OR REPLACE FUNCTION update_playlists_on_track_change()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE playlists
    SET time = calculate_playlist_time(tracks)
    WHERE NEW.id = ANY(tracks);

    UPDATE playlists
    SET allowed_time = calculate_playlist_time(allowed_tracks)
    WHERE NEW.id = ANY(allowed_tracks);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- name: InitTrackUpdate :exec
DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1 FROM pg_trigger
            WHERE tgname = 'track_update_trigger'
        ) THEN
            CREATE TRIGGER track_update_trigger
                AFTER INSERT OR UPDATE OF length ON tracks
                FOR EACH ROW
            EXECUTE FUNCTION update_playlists_on_track_change();
        END IF;
    END
$$;

-- INIT DATABASE TABLES / SCHEMA END


-- =================================


-- PLAYLISTS CRUD BEGIN

-- name: CreatePlaylist :exec
INSERT INTO playlists (id, title, thumbnail, tracks, allowed_tracks, type)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: EditPlaylist :exec
UPDATE playlists
SET
    title = COALESCE($2, title),
    thumbnail = COALESCE($3, thumbnail),
    tracks = COALESCE($4, tracks),
    allowed_tracks = COALESCE($5, allowed_tracks),
    type = COALESCE($6, type)
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

-- name: GetRole :one
SELECT playlist_id FROM playlist_permissions
WHERE user_id = $1 AND role = $2;

-- ROLES CRUD END
