package resolver

import (
	"strconv"

	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/anizip"
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
		Cover:         m.CoverImage.Large,
		Banner:        m.BannerImage,
		Score:         float64(m.AverageScore) / 10,
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
		Number:   az.EpisodeNumber,
		Title:    title,
		AirDate:  az.AirDateUtc,
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
