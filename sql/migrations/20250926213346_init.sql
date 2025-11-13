-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TYPE playlist_type AS ENUM('spotify', 'youtube', 'yandex');
CREATE TYPE playlist_role AS ENUM('viewer', 'moderator', 'owner');

CREATE TABLE IF NOT EXISTS playlists (
    id TEXT NOT NULL PRIMARY KEY UNIQUE,
    title TEXT NOT NULL,
    thumbnail TEXT NOT NULL,
    type playlist_type NOT NULL,
    external_id TEXT NOT NULL,
    telegram_id BIGINT NOT NULL,
    tracks TEXT[] DEFAULT '{}',
    allowed_tracks TEXT[] DEFAULT '{}',
    count INTEGER GENERATED ALWAYS AS (COALESCE(array_length(tracks, 1), 0)) STORED,
    allowed_count INTEGER GENERATED ALWAYS AS (COALESCE(array_length(allowed_tracks, 1), 0)) STORED,
    time INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tracks (
     id TEXT NOT NULL PRIMARY KEY,
     title TEXT NOT NULL,
     authors TEXT NOT NULL,
     thumbnail TEXT NOT NULL,
     length INTEGER NOT NULL,
     explicit BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS users (
      id BIGINT NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS playlist_permissions (
      playlist_id TEXT NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
      user_id BIGINT NOT NULL REFERENCES users(id),
      role playlist_role NOT NULL,
      PRIMARY KEY (playlist_id, user_id)
);

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

CREATE OR REPLACE FUNCTION update_playlist_time()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.time = calculate_playlist_time(NEW.allowed_tracks);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_playlist_on_track_change()
    RETURNS TRIGGER AS $$
BEGIN
    UPDATE playlists
    SET time = calculate_playlist_time(tracks)
    WHERE NEW.id = ANY(tracks);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE INDEX IF NOT EXISTS idx_tracks_id ON tracks (id);

CREATE INDEX IF NOT EXISTS idx_playlists_tracks ON playlists USING GIN(tracks);

CREATE INDEX IF NOT EXISTS idx_playlists_allowed_tracks ON playlists USING GIN(allowed_tracks);

CREATE INDEX IF NOT EXISTS idx_permissions_user ON playlist_permissions (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP FUNCTION IF EXISTS update_playlist_on_track_change() CASCADE;
DROP FUNCTION IF EXISTS update_playlist_time() CASCADE;
DROP FUNCTION IF EXISTS calculate_playlist_time(TEXT[]) CASCADE;

DROP INDEX IF EXISTS idx_permissions_user;
DROP INDEX IF EXISTS idx_playlists_allowed_tracks;
DROP INDEX IF EXISTS idx_playlists_tracks;
DROP INDEX IF EXISTS idx_tracks_id;

DROP TABLE IF EXISTS playlist_permissions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS playlists;

DROP type playlist_type;
DROP type playlist_role;
-- +goose StatementEnd
