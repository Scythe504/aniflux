package anilist

type STATUS string

const (
	StatusFinished       STATUS = "FINISHED"
	StatusReleasing      STATUS = "RELEASING"
	StatusNotYetReleased STATUS = "NOT_YET_RELEASED"
	StatusCancelled      STATUS = "CANCELLED"
	StatusHiatus         STATUS = "HIATUS"
)

type SEASON string

type ORIGIN_COUNTRY string

const (
	JP ORIGIN_COUNTRY = "JP"
)

const (
	FALL   SEASON = "FALL"
	SPRING SEASON = "SPRING"
	WINTER SEASON = "WINTER"
	SUMMER SEASON = "SUMMER"
)

type Genre string

const (
	GenreAction        Genre = "Action"
	GenreAdventure     Genre = "Adventure"
	GenreComedy        Genre = "Comedy"
	GenreDrama         Genre = "Drama"
	GenreEcchi         Genre = "Ecchi"
	GenreFantasy       Genre = "Fantasy"
	GenreHentai        Genre = "Hentai"
	GenreHorror        Genre = "Horror"
	GenreMahouShoujo   Genre = "Mahou Shoujo"
	GenreMecha         Genre = "Mecha"
	GenreMusic         Genre = "Music"
	GenreMystery       Genre = "Mystery"
	GenrePsychological Genre = "Psychological"
	GenreRomance       Genre = "Romance"
	GenreSciFi         Genre = "Sci-Fi"
	GenreSliceOfLife   Genre = "Slice of Life"
	GenreSports        Genre = "Sports"
	GenreSupernatural  Genre = "Supernatural"
	GenreThriller      Genre = "Thriller"
)

type Media struct {
	ID                int                `json:"id"`
	Episodes          *int               `json:"episodes"`
	BannerImage       string             `json:"bannerImage"`
	Season            SEASON             `json:"season"`
	Popularity        int                `json:"popularity"`
	Duration          int                `json:"duration"`
	CoverImage        CoverImage         `json:"coverImage"`
	CountryOfOrigin   ORIGIN_COUNTRY     `json:"countryOfOrigin"`
	AverageScore      int                `json:"averageScore"`
	Format            string             `json:"format"`
	Description string `json:"description"`
	Genres            []Genre            `json:"genres"`
	Title             Title              `json:"title"`
	NextAiringEpisode *NextAiringEpisode `json:"nextAiringEpisode"`
	Status            STATUS             `json:"status"`
	SeasonYear        int                `json:"seasonYear"`
}

type CoverImage struct {
	Color      string `json:"color,omitempty"`
	ExtraLarge string `json:"string,omitempty"`
	Large      string `json:"large,omitempty"`
	Medium     string `json:"medium,omitempty"`
}

type Title struct {
	English       string `json:"english"`
	Native        string `json:"native"`
	Romaji        string `json:"romaji"`
	UserPreferred string `json:"userPreffered"`
}

type NextAiringEpisode struct {
	AiringAt        int64 `json:"airingAt"`
	Episode         int64 `json:"episode"`
	Id              int   `json:"id"`
	MediaId         int   `json:"mediaId"`
	TimeUntilAiring int64 `json:"timeUntilAiring"`
}
