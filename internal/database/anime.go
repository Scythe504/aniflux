package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AnimeRecord struct {
	ID            string  `db:"id"`
	Type          string  `db:"type"`
	Title         string  `db:"title"`
	OriginalTitle string  `db:"original_title"`
	Cover         string  `db:"cover"`
	Banner        string  `db:"banner"`
	Description   string  `db:"description"`
	Score         float64 `db:"score"`
	Genres        string  `db:"genres"` // JSON string, unmarshal after
	Status        string  `db:"status"`
	Season        string  `db:"season"`
	SeasonYear    int     `db:"season_year"`
	TotalEpisodes *int    `db:"total_episodes"`
	Duration      *int    `db:"duration"`
	NextAiringEp  *int    `db:"next_airing_episode"`
	NextAiringAt  *int64  `db:"next_airing_at"`
	UpdatedAt     int64   `db:"updated_at"`
	CreatedAt     int64   `db:"created_at"`
}

func (s *service) UpsertAnime(ctx context.Context, a AnimeRecord) error {
	genres, err := json.Marshal(a.Genres)
	if err != nil {
		return fmt.Errorf("failed to marshal genres: %w", err)
	}

	stmt := `
		INSERT INTO anime (
			id, type, title, original_title, cover, banner, description,
			score, genres, status, season, season_year, total_episodes,
			duration, next_airing_episode, next_airing_at, updated_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			original_title = excluded.original_title,
			cover = excluded.cover,
			banner = excluded.banner,
			description = excluded.description,
			score = excluded.score,
			genres = excluded.genres,
			status = excluded.status,
			season = excluded.season,
			season_year = excluded.season_year,
			total_episodes = excluded.total_episodes,
			duration = excluded.duration,
			next_airing_episode = excluded.next_airing_episode,
			next_airing_at = excluded.next_airing_at,
			updated_at = excluded.updated_at
	`

	_, err = s.db.ExecContext(ctx, stmt,
		a.ID, a.Type, a.Title, a.OriginalTitle, a.Cover, a.Banner, a.Description,
		a.Score, string(genres), a.Status, a.Season, a.SeasonYear, a.TotalEpisodes,
		a.Duration, a.NextAiringEp, a.NextAiringAt,
		time.Now().UnixMilli(), time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert anime: %w", err)
	}

	return nil
}

func (s *service) GetCurrentAiring(ctx context.Context, page, perPage int) ([]AnimeRecord, error) {
	var results []AnimeRecord
	err := s.db.SelectContext(ctx, &results, `
        SELECT * FROM anime
        WHERE status = 'RELEASING'
        ORDER BY updated_at DESC
        LIMIT ? OFFSET ?
    `, perPage, (page-1)*perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get current airing: %w", err)
	}
	return results, nil
}
