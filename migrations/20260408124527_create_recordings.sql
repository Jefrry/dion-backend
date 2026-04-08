-- +goose Up
CREATE TABLE recordings (
    id            SERIAL PRIMARY KEY,
    title         VARCHAR(500) NOT NULL,
    slug          VARCHAR(500) NOT NULL UNIQUE,
    description   TEXT,
    artist_id     INT NOT NULL REFERENCES artists(id),
    concert_date  DATE,
    youtube_id    VARCHAR(20),
    external_url  TEXT,
    thumbnail_url TEXT,
    status        VARCHAR(20) DEFAULT 'pending',
    submitted_at  TIMESTAMPTZ DEFAULT NOW(),
    moderated_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS recordings;