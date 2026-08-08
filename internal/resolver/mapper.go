package resolver

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/anizip"
	"github.com/scythe504/aniflux/internal/database"
	"github.com/scythe504/aniflux/internal/sources"
)

func toMedia(m anilist.Media) Media {
	title := m.Title.English
	if title == "" {
		title = m.Title.Romaji
	}
	return Media{
		ID:            strconv.Itoa(m.ID),
		Type:          MediaTypeAnime,
		Title:         title,
		OriginalTitle: m.Title.Romaji,
		Description:   m.Description,
		Cover:         m.CoverImage.Large,
		Banner:        m.BannerImage,
		Score:         float64(m.AverageScore),
		Genres:        toGenres(m.Genres),
		Status:        toStatus(m.Status),
		Season:        Season(m.Season),
		SeasonYear:    m.SeasonYear,
		TotalEpisodes: m.Episodes,
		Duration:      &m.Duration,
		NextAiring:    toAiring(m.NextAiringEpisode),
	}
}

func toGenres(genres []anilist.Genre) []string {
	mediaGenres := make([]string, 0)

	for _, genre := range genres {
		mediaGenres = append(mediaGenres, string(genre))
	}

	return mediaGenres
}

func toStatus(status anilist.STATUS) Status {
	switch status {
	case anilist.StatusReleasing:
		return StatusReleasing
	case anilist.StatusCancelled:
		return StatusCancelled
	case anilist.StatusHiatus:
		return StatusNotYetAired
	case anilist.StatusFinished:
		return StatusFinished
	default:
		return StatusNotYetAired
	}
}

func toAiring(n *anilist.NextAiringEpisode) *Airing {
	if n == nil {
		return nil
	}

	return &Airing{
		Episode:  int(n.Episode),
		AiringAt: n.AiringAt,
	}
}

func toEpisode(az *anizip.Episode) Episode {
	title := az.Title.EN
	if title == "" {
		title = az.Title.XJAT
	}

	return Episode{
		Id:       az.AniDbEid,
		Number:   strconv.Itoa(az.EpisodeNumber),
		Title:    title,
		AirDate:  parseAirDate(az.AirDateUtc),
		Overview: az.Overview,
		Image:    az.Image,
	}
}

func toSource(s *sources.TorznabItem) Source {

	return Source{
		Title:     s.Title,
		MagnetURI: s.MagnetURI,
		InfoHash:  s.InfoHash,
		Seeders:   s.Seeders,
		Leechers:  s.Leechers,
		Size:      s.Size,
	}
}

func toAnimeRecord(m *Media) database.AnimeRecord {
	marshaledGenres, _ := json.Marshal(m.Genres)

	var nextAiringEp *int
	var nextAiringAt *int64
	if m.NextAiring != nil {
		nextAiringEp = &m.NextAiring.Episode
		nextAiringAt = &m.NextAiring.AiringAt
	}

	return database.AnimeRecord{
		ID:            m.ID,
		Type:          string(MediaTypeAnime),
		Title:         m.Title,
		OriginalTitle: m.OriginalTitle,
		Cover:         m.Cover,
		Banner:        m.Banner,
		Description:   m.Description,
		Score:         m.Score,
		Genres:        string(marshaledGenres),
		Status:        string(m.Status),
		Season:        string(m.Season),
		SeasonYear:    m.SeasonYear,
		TotalEpisodes: m.TotalEpisodes,
		Duration:      m.Duration,
		NextAiringEp:  nextAiringEp,
		NextAiringAt:  nextAiringAt,
		UpdatedAt:     time.Now().UnixMilli(),
		CreatedAt:     time.Now().UnixMilli(),
	}
}

func fromAnimeRecord(ar *database.AnimeRecord) Media {
	var genres []string
	err := json.Unmarshal([]byte(ar.Genres), &genres)
	if err != nil {
		// Attempt to recover double-marshaled JSON string
		var rawString string
		if json.Unmarshal([]byte(ar.Genres), &rawString) == nil {
			if json.Unmarshal([]byte(rawString), &genres) == nil {
				err = nil
			}
		}
		if err != nil {
			genres = strings.Split(ar.Genres, " ")
			log.Println("Failed to unmarshal genres:", ar.Genres)
		}
	}

	var nextAiring *Airing
	if ar.NextAiringEp != nil && ar.NextAiringAt != nil {
		nextAiring = &Airing{
			Episode:  *ar.NextAiringEp,
			AiringAt: *ar.NextAiringAt,
		}
	}

	return Media{
		ID:            ar.ID,
		Type:          MediaType(ar.Type),
		Title:         ar.Title,
		OriginalTitle: ar.OriginalTitle,
		Cover:         ar.Cover,
		Banner:        ar.Banner,
		Description:   ar.Description,
		Score:         ar.Score,
		Genres:        genres,
		Status:        Status(ar.Status),
		Season:        Season(ar.Season),
		SeasonYear:    ar.SeasonYear,
		TotalEpisodes: ar.TotalEpisodes,
		Duration:      ar.Duration,
		NextAiring:    nextAiring,
	}
}

func fromAiringRecord(ar *database.AiringRecord) Episode {
	title := ar.Title
	if title == "" {
		title = ar.OriginalTitle
	}
	animeIDInt, _ := strconv.Atoi(ar.AnimeID)
	return Episode{
		Id:          animeIDInt,
		Number:      strconv.Itoa(ar.Episode),
		Title:       title,
		AirDate:     ar.AiringAt * 1000,
		Overview:    "",
		Image:       ar.Cover,
		ScheduleDay: ar.ScheduleDay,
		Timing:      ar.Timing,
	}
}

func parseAirDate(dateStr string) int64 {
	if dateStr == "" {
		return 0
	}
	// Try parsing RFC3339 format (e.g. 2024-05-26T15:30:00Z)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t.UnixMilli()
	}
	// Try parsing "2006-01-02" date format
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t.UnixMilli()
	}
	// Try parsing "2006-01-02 15:04:05" layout
	t, err = time.Parse("2006-01-02 15:04:05", dateStr)
	if err == nil {
		return t.UnixMilli()
	}
	return 0
}