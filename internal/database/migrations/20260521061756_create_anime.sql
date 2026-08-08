-- +goose Up
CREATE TABLE IF NOT EXISTS anime (
    id TEXT PRIMARY KEY NOT NULL,
    type TEXT NOT NULL DEFAULT 'anime',
    title TEXT NOT NULL,
    original_title TEXT,
    cover TEXT,
    banner TEXT,
    description TEXT,
    score REAL,
    genres TEXT, -- JSON array stored as string
    status TEXT NOT NULL,
    season TEXT,
    season_year INTEGER,
    total_episodes INTEGER,
    duration INTEGER,
    next_airing_episode INTEGER,
    next_airing_at BIGINT,
    updated_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_anime_status ON anime(status);
CREATE INDEX IF NOT EXISTS idx_anime_season ON anime(season, season_year);
CREATE INDEX IF NOT EXISTS idx_anime_updated ON anime(updated_at);

-- +goose Down
DROP TABLE IF EXISTS anime;