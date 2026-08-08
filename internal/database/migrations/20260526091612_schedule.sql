-- +goose Up
CREATE TABLE IF NOT EXISTS airing_schedule (
    id SERIAL PRIMARY KEY,
    anime_id TEXT NOT NULL REFERENCES anime(id),
    episode INTEGER NOT NULL,
    airing_at BIGINT NOT NULL,
    time_until_airing BIGINT,
    updated_at BIGINT NOT NULL,
    UNIQUE(anime_id, episode)
);

CREATE INDEX IF NOT EXISTS idx_airing_anime ON airing_schedule(anime_id);
CREATE INDEX IF NOT EXISTS idx_airing_at ON airing_schedule(airing_at);

-- +goose Down
DROP TABLE IF EXISTS airing_schedule;