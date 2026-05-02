-- +goose Up
ALTER TABLE recordings
    ADD COLUMN artist_name VARCHAR(255);

UPDATE recordings r
SET artist_name = a.name
FROM artists a
WHERE r.artist_id = a.id;

ALTER TABLE recordings
    ALTER COLUMN artist_name SET NOT NULL,
    ALTER COLUMN artist_id DROP NOT NULL;

-- +goose Down
DELETE FROM recordings
WHERE artist_id IS NULL;

ALTER TABLE recordings
    ALTER COLUMN artist_id SET NOT NULL,
    DROP COLUMN artist_name;
