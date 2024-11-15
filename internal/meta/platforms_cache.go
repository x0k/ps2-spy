package meta

import (
	"context"
	"errors"

	"github.com/x0k/ps2-spy/internal/lib/cache"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrPlatformCacheNotFound = errors.New("cache for platform not found")

type PlatformsCache[K comparable, T any] struct {
	platformsCaches map[ps2_platforms.Platform]cache.Multi[K, T]
}

func NewPlatformsCache[K comparable, T any](
	caches map[ps2_platforms.Platform]cache.Multi[K, T],
) *PlatformsCache[K, T] {
	return &PlatformsCache[K, T]{
		platformsCaches: caches,
	}
}

func (c *PlatformsCache[K, T]) Get(ctx context.Context, query PlatformQuery[[]K]) (map[K]T, bool) {
	cache, ok := c.platformsCaches[query.Platform]
	if !ok {
		return nil, false
	}
	return cache.Get(ctx, query.Value)
}

func (c *PlatformsCache[K, T]) Add(ctx context.Context, query PlatformQuery[[]K], values map[K]T) error {
	cache, ok := c.platformsCaches[query.Platform]
	if !ok {
		return ErrPlatformCacheNotFound
	}
	return cache.Add(ctx, values)
}
