package resolver

import (
	"context"
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
	PLuginInfo() string
	Trending(ctx context.Context, page, perPage int) ([]Media, error)
	GetMedia(ctx context.Context, id int) (*Media, error)
	GetEpisodes(ctx context.Context, id, page, perPage int) (*EpisodeList, error)
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
func (rs *resolver) PLuginInfo() string { return "aniflux" }

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

	if cached, ok := rs.cache.Load(id); ok {
		resp = cached.(*anizip.Resp)
	} else {
		ep, err := rs.anizip.FetchAnizipData(ctx, anizip.AnilistID, id)
		if err != nil {
			return nil, err
		}
		rs.cache.Store(id, ep)
		resp = ep
	}

	episodes := make([]Episode, 0)
	specials := make([]Episode, 0)
	for key, val := range resp.Episodes {
		ep := toEpisode(&val)
		if _, err := strconv.Atoi(key); err != nil {
			// non-numeric key = special/OVA/credit/trailer
			specials = append(specials, ep)
		} else {
			episodes = append(episodes, ep)
		}
	}

	sort.Slice(episodes, func(i, j int) bool {
		return episodes[i].Number < episodes[j].Number
	})

	return &EpisodeList{
		Episodes: utils.Paginate(episodes, page, perPage),
		Specials: specials,
	}, nil
}
