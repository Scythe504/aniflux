package resolver

import (
	"context"

	"github.com/scythe504/aniflux/internal/anilist"
	"github.com/scythe504/aniflux/internal/anizip"
	"github.com/scythe504/aniflux/internal/database"
	"github.com/scythe504/aniflux/internal/sources"
)

type Resolver struct {
	db      database.Service
	anilist *anilist.Client
	anizip  *anizip.Client
	sources *sources.Client
}

// Returns plugin details
func (rs *Resolver) PLuginInfo() string { return "aniflux" }

func (rs *Resolver) Trending(ctx context.Context, page, perPage int) ([]Media, error) {
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