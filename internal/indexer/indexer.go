package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/database"
)

// MapAnilistMediaToAnimeRecord maps anilist.Media to database.AnimeRecord
func MapAnilistMediaToAnimeRecord(m anilist.Media) database.AnimeRecord {
	title := m.Title.English
	if title == "" {
		title = m.Title.Romaji
	}

	genresBytes, _ := json.Marshal(m.Genres)
	genresStr := string(genresBytes)

	var status string
	switch m.Status {
	case anilist.StatusReleasing:
		status = "RELEASING"
	case anilist.StatusCancelled:
		status = "CANCELLED"
	case anilist.StatusHiatus:
		status = "NOT_YET_AIRED"
	case anilist.StatusFinished:
		status = "FINISHED"
	default:
		status = "NOT_YET_AIRED"
	}

	var nextEp *int
	var nextAt *int64
	if m.NextAiringEpisode != nil {
		ep := int(m.NextAiringEpisode.Episode)
		nextEp = &ep
		nextAt = &m.NextAiringEpisode.AiringAt
	}

	duration := m.Duration

	return database.AnimeRecord{
		ID:            strconv.Itoa(m.ID),
		Type:          "anime",
		Title:         title,
		OriginalTitle: m.Title.Romaji,
		Cover:         m.CoverImage.Large,
		Banner:        m.BannerImage,
		Description:   m.Description,
		Score:         float64(m.AverageScore),
		Genres:        genresStr,
		Status:        status,
		Season:        string(m.Season),
		SeasonYear:    m.SeasonYear,
		TotalEpisodes: m.Episodes,
		Duration:      &duration,
		NextAiringEp:  nextEp,
		NextAiringAt:  nextAt,
		UpdatedAt:     time.Now().UnixMilli(),
		CreatedAt:     time.Now().UnixMilli(),
	}
}

// UpdateWeeklySchedule fetches upcoming broadcasts for the week and syncs them to the DB
func UpdateWeeklySchedule(ctx context.Context, db database.Service, client *anilist.Client) error {
	jstLocation := time.FixedZone("JST", 9*60*60)
	nowInJst := time.Now().In(jstLocation)

	var startTimeJst time.Time
	var endTimeJst time.Time

	if nowInJst.Weekday() == time.Monday {
		// Monday JP execution: Start of Monday JST 00:00:00 to 7 days later
		startTimeJst = time.Date(nowInJst.Year(), nowInJst.Month(), nowInJst.Day(), 0, 0, 0, 0, jstLocation)
		endTimeJst = startTimeJst.Add(7 * 24 * time.Hour)
	} else {
		// Non-Monday JP execution: Start of today JST 00:00:00 to 24 hours later
		startTimeJst = time.Date(nowInJst.Year(), nowInJst.Month(), nowInJst.Day(), 0, 0, 0, 0, jstLocation)
		endTimeJst = startTimeJst.Add(24 * time.Hour)
	}

	airingAtGreater := int(startTimeJst.Unix())
	airingAtLesser := int(endTimeJst.Unix())

	page := 1
	perPage := 50

	for {
		log.Printf("Fetching airing schedule page %d...", page)
		schedules, err := client.FetchAiringSchedule(ctx, page, perPage, airingAtLesser, airingAtGreater)
		if err != nil {
			return err
		}

		if len(schedules) == 0 {
			break
		}

		for _, s := range schedules {
			// Ensure the anime is indexed in our database
			animeRec := MapAnilistMediaToAnimeRecord(s.Media)
			if err := db.UpsertAnime(ctx, animeRec); err != nil {
				log.Printf("Failed to upsert anime %s: %v", animeRec.Title, err)
				continue
			}

			// Upsert the airing schedule details
			title := s.Media.Title.English
			if title == "" {
				title = s.Media.Title.Romaji
			}

			airingRec := database.AiringRecord{
				AnimeID:         animeRec.ID,
				Episode:         s.Episode,
				AiringAt:        s.AiringAt,
				TimeUntilAiring: s.TimeUntilAiring,
				UpdatedAt:       time.Now().UnixMilli(),
				Title:           title,
				OriginalTitle:   s.Media.Title.Romaji,
				Cover:           s.Media.CoverImage.Large,
			}

			if err := db.UpsertAiringSchedule(airingRec); err != nil {
				log.Printf("Failed to upsert airing record for anime %s episode %d: %v", animeRec.Title, s.Episode, err)
				continue
			}
		}

		if len(schedules) < perPage {
			break
		}
		page++
	}

	log.Println("Weekly schedule update finished successfully.")
	return nil
}

// UpdateSeasonalMedia loops page by page to sync all anime for a specific season and year
func UpdateSeasonalMedia(ctx context.Context, db database.Service, client *anilist.Client, seasonStr string, year int) error {
	var season anilist.SEASON
	switch strings.ToUpper(seasonStr) {
	case "WINTER":
		season = anilist.WINTER
	case "SPRING":
		season = anilist.SPRING
	case "SUMMER":
		season = anilist.SUMMER
	case "FALL":
		season = anilist.FALL
	default:
		return fmt.Errorf("invalid season value: %s", seasonStr)
	}

	page := 1
	perPage := 50

	for {
		log.Printf("Fetching seasonal media page %d for %s %d...", page, season, year)
		mediaItems, err := client.FetchMediaBySeason(ctx, page, perPage, &season, &year)
		if err != nil {
			return err
		}

		if len(mediaItems) == 0 {
			break
		}

		records := make([]database.AnimeRecord, len(mediaItems))
		for idx, m := range mediaItems {
			records[idx] = MapAnilistMediaToAnimeRecord(m)
		}

		if err := db.BulkUpsertAnime(ctx, records); err != nil {
			return fmt.Errorf("failed to bulk upsert seasonal media on page %d: %w", page, err)
		}

		if len(mediaItems) < perPage {
			break
		}
		page++
	}

	log.Printf("Seasonal index update finished successfully for %s %d.", season, year)
	return nil
}

// UpdateMediaEntry fetches metadata for a single anime by its AniList ID and upserts it in the DB
func UpdateMediaEntry(ctx context.Context, db database.Service, client *anilist.Client, anilistID int) error {
	log.Printf("Fetching media metadata for AniList ID %d...", anilistID)
	m, err := client.FetchAnilistMedia(ctx, anilistID)
	if err != nil {
		return err
	}

	animeRec := MapAnilistMediaToAnimeRecord(*m)
	if err := db.UpsertAnime(ctx, animeRec); err != nil {
		return err
	}

	log.Printf("Successfully indexed anime entry: %s (ID: %d)", animeRec.Title, anilistID)
	return nil
}
