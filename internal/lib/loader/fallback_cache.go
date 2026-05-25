package loader

import (
	"context"
	"log/slog"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type SimpleFallbackCache[T any] struct {
	fallbacks *containers.Fallbacks[Simple[T]]
	Load      Keyed[string, T]
}

func NewSimpleFallbackCache[T any](
	log *slog.Logger,
	loaders map[string]Simple[T],
	priority []string,
	cacheSize int,
) *SimpleFallbackCache[T] {
	fallbacks := containers.NewFallbacks(log, loaders, priority, time.Hour)
	fallbackLoader := NewFallback(fallbacks)
	cached := WithKeyedCache(
		log,
		func(ctx context.Context, provider string) (T, error) {
			if l, ok := loaders[provider]; ok {
				return l(ctx)
			}
			return fallbackLoader(ctx)
		},
		memory.NewKeyedExpirableCache(
			expirable.NewLRU[string, T](
				cacheSize,
				nil,
				time.Minute,
			),
		),
	)
	return &SimpleFallbackCache[T]{
		fallbacks: fallbacks,
		Load:      cached,
	}
}

func (c *SimpleFallbackCache[T]) Start(ctx context.Context) {
	c.fallbacks.Start(ctx)
}

type KeyedFallbackCache[K comparable, T any] struct {
	fallbacks *containers.Fallbacks[Keyed[K, T]]
	Load      Queried[Query[K], T]
}

func NewKeyedFallbackCache[K comparable, T any](
	log *slog.Logger,
	loaders map[string]Keyed[K, T],
	priority []string,
	cacheSize int,
) *KeyedFallbackCache[K, T] {
	fallbacks := containers.NewFallbacks(log, loaders, priority, time.Hour)
	fallbackLoader := NewKeyedFallback(fallbacks)
	cached := WithQueriedCache(
		log,
		func(ctx context.Context, query Query[K]) (T, error) {
			if l, ok := loaders[query.Provider]; ok {
				return l(ctx, query.Key)
			}
			return fallbackLoader(ctx, query.Key)
		},
		memory.NewKeyedExpirableCache(
			expirable.NewLRU[Query[K], T](
				cacheSize,
				nil,
				time.Minute,
			),
		),
	)
	return &KeyedFallbackCache[K, T]{
		fallbacks: fallbacks,
		Load:      cached,
	}
}

func (c *KeyedFallbackCache[K, T]) Start(ctx context.Context) {
	c.fallbacks.Start(ctx)
}
