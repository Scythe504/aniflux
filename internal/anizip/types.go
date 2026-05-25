package anizip

type MappingName string

const (
	AnizipID      MappingName = "anizip_id"
	AnimePlanetID MappingName = "animeplanet_id"
	KitsuID       MappingName = "kitsu_id"
	KitsuIDKey    MappingName = "kitsu_id"
	MalID         MappingName = "mal_id"
	AnilistID     MappingName = "anilist_id"
	AnisearchID   MappingName = "anisearch_id"
	AnidbID       MappingName = "anidb_id"
	LivechartID   MappingName = "livechart_id"
	ThetvdbID     MappingName = "thetvdb_id"
	ImdbID        MappingName = "imdb_id"
	ThemoviedbID  MappingName = "themoviedb_id"
)

type Episode struct {
	TvDbId                int64      `json:"tvdbId"`
	TvDbShowId            int64      `json:"tvdbShowId"`
	SeasonNumber          int        `json:"seasonNumber"`
	EpisodeNumber         int        `json:"episodeNumber"`
	AbsoluteEpisodeNumber int        `json:"absoluteEpisodeNumber"`
	AirDate               string     `json:"airDate"`
	AirDateUtc            string     `json:"airDateUtc"`
	Overview              string     `json:"overview"`
	Title                 AnidbTitle `json:"title"`
	Image                 string     `json:"image"`
	Episode               string     `json:"episode"`
	AniDbEid              int        `json:"anidbEid"`
	Length                int        `json:"length"`
	Rating                string     `json:"rating"`
	Summary               string     `json:"summary"`
}

type AnidbTitle struct {
	JA   string `json:"ja"`
	EN   string `json:"en"`
	XJAT string `json:"x-jat"`
}

type ImageCoverType string

const (
	BANNER    ImageCoverType = "Banner"
	POSTER    ImageCoverType = "Poster"
	FANART    ImageCoverType = "Fanart"
	CLEARLOGO ImageCoverType = "Clearlogo"
)

type Image struct {
	CoverType ImageCoverType `json:"coverType"`
	Url       string         `json:"url"`
}

type Mappings struct {
	AnimePlanetID string  `json:"animeplanet_id"`
	KitsuID       int     `json:"kitsu_id"`
	MalID         int     `json:"mal_id"`
	Type          string  `json:"type"`
	AnilistID     int     `json:"anilist_id"`
	AnisearchID   int     `json:"anisearch_id"`
	AnidbID       int     `json:"anidb_id"`
	NotifymoeID   *string `json:"notifymoe_id"`
	LivechartID   int     `json:"livechart_id"`
	ThetvdbID     int     `json:"thetvdb_id"`
	ImdbID        string  `json:"imdb_id"`
	ThemoviedbID  string  `json:"themoviedb_id"`
}

type Resp struct {
	Titles       AnidbTitle         `json:"titles"`
	Episodes     map[string]Episode `json:"episodes"`
	EpisodeCount uint               `json:"episodeCount"`
	SpecialCount uint               `json:"specialCount"`
	Images       []Image            `json:"images"`
	Mappings     Mappings           `json:"mappings"`
}
