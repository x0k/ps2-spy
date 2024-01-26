package outfits_cache

import (
	"context"
	"errors"

	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrPlatformCacheNotFound = errors.New("platform cache not found")

type MultiPlatform struct {
	platformsCaches map[platforms.Platform]containers.ContextMultiKeyedCache[ps2.OutfitId, ps2.Outfit]
}

func NewMultiPlatform(
	caches map[platforms.Platform]containers.ContextMultiKeyedCache[ps2.OutfitId, ps2.Outfit],
) *MultiPlatform {
	return &MultiPlatform{
		platformsCaches: caches,
	}
}

func (c *MultiPlatform) Get(ctx context.Context, query meta.PlatformQuery[ps2.OutfitId]) (map[ps2.OutfitId]ps2.Outfit, bool) {
	cache, ok := c.platformsCaches[query.Platform]
	if !ok {
		return nil, false
	}
	return cache.Get(ctx, query.Items)
}

func (c *MultiPlatform) Add(ctx context.Context, query meta.PlatformQuery[ps2.OutfitId], values map[ps2.OutfitId]ps2.Outfit) error {
	cache, ok := c.platformsCaches[query.Platform]
	if !ok {
		return ErrPlatformCacheNotFound
	}
	return cache.Add(ctx, values)
}
