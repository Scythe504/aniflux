package database

import (
	"context"
	"fmt"
	"time"

	"github.com/scythe504/aniflux/internal/utils"
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

type AiringRecord struct {
	ID              int    `db:"id"`
	AnimeID         string `db:"anime_id"`
	Episode         int    `db:"episode"`
	AiringAt        int64  `db:"airing_at"`
	TimeUntilAiring int64  `db:"time_until_airing"`
	UpdatedAt       int64  `db:"updated_at"`
	Title           string `db:"title"`
	OriginalTitle   string `db:"original_title"`
	Cover           string `db:"cover"`
	ScheduleDay     string `db:"schedule_day"`
	Timing          string `db:"timing"`
}

func (s *service) UpsertAiringSchedule(a AiringRecord) error {
	scheduleDay, timing := utils.Get30HourSchedule(a.AiringAt)

	query := `
        INSERT INTO airing_schedule (anime_id, episode, airing_at, time_until_airing, updated_at, schedule_day, timing)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT(anime_id, episode) DO UPDATE SET
            airing_at = excluded.airing_at,
            time_until_airing = excluded.time_until_airing,
            updated_at = excluded.updated_at,
            schedule_day = excluded.schedule_day,
            timing = excluded.timing
    `
	_, err := s.pool.Exec(context.Background(), query, a.AnimeID, a.Episode, a.AiringAt, a.TimeUntilAiring, time.Now().UnixMilli(), scheduleDay, timing)
	if err != nil {
		return fmt.Errorf("failed to upsert airing schedule: %w", err)
	}
	return nil
}

func (s *service) GetWeeklySchedule() ([]AiringRecord, error) {
	now := time.Now().Unix()
	weekEnd := now + (7 * 24 * 60 * 60)

	var results []AiringRecord
	query := `
        SELECT 
            asch.id,
            asch.anime_id,
            asch.episode,
            asch.airing_at,
            asch.time_until_airing,
            asch.updated_at,
            COALESCE(a.title, '') AS title,
            COALESCE(a.original_title, '') AS original_title,
            COALESCE(a.cover, '') AS cover,
            asch.schedule_day,
            asch.timing
        FROM airing_schedule asch
        LEFT JOIN anime a ON asch.anime_id = a.id
        WHERE asch.airing_at BETWEEN $1 AND $2
        ORDER BY asch.airing_at ASC
    `
	rows, err := s.pool.Query(context.Background(), query, now, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly schedule: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r AiringRecord
		err := rows.Scan(
			&r.ID,
			&r.AnimeID,
			&r.Episode,
			&r.AiringAt,
			&r.TimeUntilAiring,
			&r.UpdatedAt,
			&r.Title,
			&r.OriginalTitle,
			&r.Cover,
			&r.ScheduleDay,
			&r.Timing,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan weekly schedule row: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("weekly schedule rows error: %w", err)
	}

	return results, nil
}

// UpsertAnime inserts a new anime record or updates key fields on conflict.
// The updated_at field is only bumped when the next_airing_at progresses (indicating
// a new episode has broadcasted), setting it to the broadcast time of the aired episode.
func (s *service) UpsertAnime(ctx context.Context, a AnimeRecord) error {
	stmt := `
		INSERT INTO anime (
			id, type, title, original_title, cover, banner, description,
			score, genres, status, season, season_year, total_episodes,
			duration, next_airing_episode, next_airing_at, updated_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT(id) DO UPDATE SET
			cover = excluded.cover,
			banner = excluded.banner,
			score = excluded.score,
			status = excluded.status,
			total_episodes = excluded.total_episodes,
			next_airing_episode = excluded.next_airing_episode,
			next_airing_at = excluded.next_airing_at,
			updated_at = CASE 
				WHEN anime.next_airing_at IS NULL THEN excluded.updated_at
				WHEN excluded.next_airing_at IS NULL THEN anime.next_airing_at * 1000
				WHEN excluded.next_airing_at > anime.next_airing_at THEN anime.next_airing_at * 1000
				ELSE anime.updated_at 
			END
	`

	_, err := s.pool.Exec(ctx, stmt,
		a.ID, a.Type, a.Title, a.OriginalTitle, a.Cover, a.Banner, a.Description,
		a.Score, a.Genres, a.Status, a.Season, a.SeasonYear, a.TotalEpisodes,
		a.Duration, a.NextAiringEp, a.NextAiringAt,
		time.Now().UnixMilli(), time.Now().UnixMilli(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert anime: %w", err)
	}

	return nil
}

// BulkUpsertAnime updates or inserts multiple anime records inside a database transaction.
func (s *service) BulkUpsertAnime(ctx context.Context, records []AnimeRecord) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	stmt := `
		INSERT INTO anime (
			id, type, title, original_title, cover, banner, description,
			score, genres, status, season, season_year, total_episodes,
			duration, next_airing_episode, next_airing_at, updated_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT(id) DO UPDATE SET
			cover = excluded.cover,
			banner = excluded.banner,
			score = excluded.score,
			status = excluded.status,
			total_episodes = excluded.total_episodes,
			next_airing_episode = excluded.next_airing_episode,
			next_airing_at = excluded.next_airing_at,
			updated_at = CASE 
				WHEN anime.next_airing_at IS NULL THEN excluded.updated_at
				WHEN excluded.next_airing_at IS NULL THEN anime.next_airing_at * 1000
				WHEN excluded.next_airing_at > anime.next_airing_at THEN anime.next_airing_at * 1000
				ELSE anime.updated_at 
			END
	`

	for _, a := range records {
		_, err = tx.Exec(ctx, stmt,
			a.ID, a.Type, a.Title, a.OriginalTitle, a.Cover, a.Banner, a.Description,
			a.Score, a.Genres, a.Status, a.Season, a.SeasonYear, a.TotalEpisodes,
			a.Duration, a.NextAiringEp, a.NextAiringAt,
			time.Now().UnixMilli(), time.Now().UnixMilli(),
		)
		if err != nil {
			return fmt.Errorf("failed to execute insert for anime %s: %w", a.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *service) GetCurrentAiring(ctx context.Context, page, perPage int) ([]AnimeRecord, error) {
	var results []AnimeRecord
	query := `
        SELECT 
            id, type, title, original_title, cover, banner, description,
            score, genres, status, season, season_year, total_episodes,
            duration, next_airing_episode, next_airing_at, updated_at, created_at
        FROM anime
        WHERE status = 'RELEASING'
        ORDER BY updated_at DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := s.pool.Query(ctx, query, perPage, (page-1)*perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to get current airing: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a AnimeRecord
		err := rows.Scan(
			&a.ID,
			&a.Type,
			&a.Title,
			&a.OriginalTitle,
			&a.Cover,
			&a.Banner,
			&a.Description,
			&a.Score,
			&a.Genres,
			&a.Status,
			&a.Season,
			&a.SeasonYear,
			&a.TotalEpisodes,
			&a.Duration,
			&a.NextAiringEp,
			&a.NextAiringAt,
			&a.UpdatedAt,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan current airing row: %w", err)
		}
		results = append(results, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("current airing rows error: %w", err)
	}

	return results, nil
}
