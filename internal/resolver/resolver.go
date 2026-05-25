package resolver

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/anizip"
	"github.com/scythe504/aniflux/internal/database"
	"github.com/scythe504/aniflux/internal/sources"
	"github.com/scythe504/aniflux/internal/utils"
)

type Resolver interface {
	PluginInfo() string
	Trending(ctx context.Context, page, perPage int) ([]Media, error)
	GetMedia(ctx context.Context, id int) (*Media, error)
	GetEpisodes(ctx context.Context, id, page, perPage int) (*EpisodeList, error)
	GetSources(ctx context.Context, anilistId int, epNumber string) ([]Source, error)
	// GetCurrentAiring(ctx context.Context, page, perPage int) ([]Media, error)                       // basically updates with cron from db every hour, shown in homescreen
	Search(ctx context.Context, query string, page, perPage int) ([]Media, error) // search
	// GetUpcomingMedia(ctx context.Context) ([]Media, error)                                          // current week
	GetRecommendations(ctx context.Context, anilistId, page, perPage int) ([]Media, error)                       // by anilistId it generates 5 recommendations
	GetMediaBySeason(ctx context.Context, season *anilist.SEASON, year *int, page, perPage int) ([]Media, error) // lists all anime for a season (fall/... in season year)
	// could be merged but weird since both season and filter can work standalone and with together basically results union (A Union B)
	GetMediaByGenre(ctx context.Context, genre []string, page, perPage int) ([]Media, error)
}

type resolver struct {
	db      database.Service
	anilist *anilist.Client
	anizip  *anizip.Client
	sources *sources.Client
	cache   sync.Map
}

var instance *resolver

func New(db database.Service) Resolver {
	if instance != nil {
		return instance
	}

	instance = &resolver{
		db:      db,
		anilist: anilist.New(),
		anizip:  anizip.New(),
		sources: sources.New(),
	}

	return instance
}

// Returns plugin details
func (rs *resolver) PluginInfo() string { return "aniflux" }

func (rs *resolver) Trending(ctx context.Context, page, perPage int) ([]Media, error) {
	trendingMedia, err := rs.anilist.FetchAnilistTrending(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	media := make([]Media, len(trendingMedia))

	for idx, t := range trendingMedia {
		media[idx] = toMedia(t)
	}

	return media, nil
}

func (rs *resolver) GetMedia(ctx context.Context, id int) (*Media, error) {
	m, err := rs.anilist.FetchAnilistMedia(ctx, id)
	if err != nil {
		return nil, err
	}

	media := toMedia(*m)

	return &media, nil
}

func (rs *resolver) GetEpisodes(ctx context.Context, id, page, perPage int) (*EpisodeList, error) {
	var resp *anizip.Resp

	if cached, ok := rs.cache.Load(fmt.Sprintf("ep:%d", id)); ok {
		resp = cached.(*anizip.Resp)
	} else {
		ep, err := rs.anizip.FetchAnizipData(ctx, anizip.AnilistID, id)
		if err != nil {
			return nil, err
		}
		rs.cache.Store(fmt.Sprintf("ep:%d", id), ep)
		resp = ep
	}

	episodes := make([]Episode, 0)
	specials := make([]Episode, 0)

	for key, val := range resp.Episodes {
		ep := toEpisode(&val)
		ep.Number = key // Explicitly inject the raw key ("6", "S1", "OP1") into the struct

		if _, err := strconv.Atoi(key); err != nil {
			// Non-numeric key = special/OVA/credit/trailer
			specials = append(specials, ep)
		} else {
			episodes = append(episodes, ep)
		}
	}

	// Correctly sort numeric strings as real numbers
	sort.Slice(episodes, func(i, j int) bool {
		ni, _ := strconv.Atoi(episodes[i].Number)
		nj, _ := strconv.Atoi(episodes[j].Number)
		return ni < nj
	})

	// Sort specials naturally (e.g. S1, S2, S11, S111)
	sort.Slice(specials, func(i, j int) bool {
		return utils.CompareNatural(specials[i].Number, specials[j].Number)
	})

	return &EpisodeList{
		Episodes:   utils.Paginate(episodes, page, perPage),
		Specials:   specials,
		TotalCount: len(episodes),
	}, nil
}

func (rs *resolver) GetSources(ctx context.Context, anilistId int, epNumber string) ([]Source, error) {
	// 1. get anidbEid from anizip cache
	var anizipResp *anizip.Resp
	if cached, ok := rs.cache.Load(fmt.Sprintf("src:%d", anilistId)); ok {
		anizipResp = cached.(*anizip.Resp)
	} else {
		resp, err := rs.anizip.FetchAnizipData(ctx, anizip.AnilistID, anilistId)
		if err != nil {
			return nil, err
		}
		rs.cache.Store(fmt.Sprintf("src:%d", anilistId), resp)
		anizipResp = resp
	}

	// 2. find anidbEid for this episode number
	ep, ok := anizipResp.Episodes[epNumber]
	if !ok {
		return nil, fmt.Errorf("episode %s not found", epNumber)
	}

	// 3. fetch sources by anidbEid
	src, err := rs.sources.FetchSources(ctx, ep.AniDbEid)
	if err != nil {
		return nil, err
	}

	episodeSources := make([]Source, len(src))
	for idx, s := range src {
		episodeSources[idx] = toSource(&s)
	}

	sort.Slice(episodeSources, func(i, j int) bool {
		return episodeSources[i].Seeders > episodeSources[j].Seeders
	})

	return episodeSources, nil
}

// cache (db) - cron updates once every end of day
// func (rs *resolver) GetCurrentAiring(ctx context.Context, page, perPage int) ([]Media, error) {
// 	resp, err := getFromDb
// }

// always realtime - cache media entry only
func (rs *resolver) Search(ctx context.Context, query string, page, perPage int) ([]Media, error) {
	anilistMedia, err := rs.anilist.Search(ctx, page, perPage, query)
	if err != nil {
		return nil, err
	}
	// push to a worker to save the individual entries
	media := make([]Media, len(anilistMedia))

	for idx, m := range anilistMedia {
		media[idx] = toMedia(m)
	}

	return media, nil
}

// cache (db) - cron updates
// func (rs *resolver) GetUpcomingMedia(ctx context.Context) ([]Media, error) {

// }

// cache (mem - ttl 1hr) - fetch realtime
func (rs *resolver) GetRecommendations(ctx context.Context, anilistId, page, perPage int) ([]Media, error) {
	var recommendations []anilist.Media
	if val, ok := rs.cache.Load(fmt.Sprintf("rec:%d:%d", anilistId, page)); ok {
		recommendations = val.([]anilist.Media)
	} else {
		recs, err := rs.anilist.FetchRecommendations(ctx, page, perPage, anilistId)
		if err != nil {
			return nil, err
		}
		rs.cache.Store(fmt.Sprintf("rec:%d:%d", anilistId, page), recs)
		recommendations = recs
	}
	var mediaRecs = make([]Media, len(recommendations))
	for idx, rec := range recommendations {
		mediaRecs[idx] = toMedia(rec)
	}

	return mediaRecs, nil
}

// cache (db) - build as we go
func (rs *resolver) GetMediaBySeason(ctx context.Context, season *anilist.SEASON, year *int, page, perPage int) ([]Media, error) {
	anilistMedia, err := rs.anilist.FetchMediaBySeason(ctx, page, perPage, season, year)
	if err != nil {
		return nil, err
	}
	media := make([]Media, len(anilistMedia))

	for idx, m := range anilistMedia {
		media[idx] = toMedia(m)
	}

	return media, nil
}

// cache (media entry only) - build as we go
func (rs *resolver) GetMediaByGenre(ctx context.Context, genre []string, page, perPage int) ([]Media, error) {
	anilistMedia, err := rs.anilist.FetchMediaByGenre(ctx, page, perPage, genre)
	if err != nil {
		return nil, err
	}

	media := make([]Media, len(anilistMedia))

	for idx, m := range anilistMedia {
		media[idx] = toMedia(m)
	}

	return media, nil
}
