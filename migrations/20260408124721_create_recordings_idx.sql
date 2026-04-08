-- +goose Up
CREATE INDEX idx_recordings_slug ON recordings(slug);
CREATE INDEX idx_recordings_artist ON recordings(artist_id);
CREATE INDEX idx_recordings_status ON recordings(status);
CREATE INDEX idx_artists_slug ON artists(slug);

-- +goose Down
DROP INDEX IF EXISTS idx_artists_slug;
DROP INDEX IF EXISTS idx_recordings_status;
DROP INDEX IF EXISTS idx_recordings_artist;
DROP INDEX IF EXISTS idx_recordings_slug;