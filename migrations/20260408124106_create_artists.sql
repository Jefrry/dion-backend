-- +goose Up
CREATE TABLE artists (
 id          SERIAL PRIMARY KEY,
 name        VARCHAR(255) NOT NULL UNIQUE,
 slug        VARCHAR(255) NOT NULL UNIQUE,
 created_at  TIMESTAMPTZ DEFAULT NOW()
);
-- +goose Down
DROP TABLE IF EXISTS artists;
