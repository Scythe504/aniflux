package resolver

import (
	"context"

	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/anizip"
	"github.com/scythe504/aniflux/internal/database"
	"github.com/scythe504/aniflux/internal/sources"
)

type Resolver interface {
	PLuginInfo() string
	Trending(ctx context.Context, page, perPage int) ([]Media, error)
}

type resolver struct {
	db      database.Service
	anilist *anilist.Client
	anizip  *anizip.Client
	sources *sources.Client
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