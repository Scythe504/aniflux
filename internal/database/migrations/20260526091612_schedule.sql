-- +goose Up
CREATE TABLE IF NOT EXISTS airing_schedule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anime_id TEXT NOT NULL REFERENCES anime(id),
    episode INTEGER NOT NULL,
    airing_at INTEGER NOT NULL,
    time_until_airing INTEGER,
    updated_at INTEGER NOT NULL,
    UNIQUE(anime_id, episode)
);

CREATE INDEX IF NOT EXISTS idx_airing_anime ON airing_schedule(anime_id);
CREATE INDEX IF NOT EXISTS idx_airing_at ON airing_schedule(airing_at);

-- +goose Down
DROP TABLE IF EXISTS airing_schedule;