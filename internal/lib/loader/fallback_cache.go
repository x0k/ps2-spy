package loader

import (
	"context"
	"log/slog"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/containers"
)

// fallbackLastSuccessTTL controls how long the fallback chain remembers the
// last successful provider before re-probing priority order.
const fallbackLastSuccessTTL = time.Hour

// cacheEntryTTL controls how long a successful loader result lives in the LRU
// cache before being re-fetched.
const cacheEntryTTL = time.Minute

// SimpleFallbackCache wraps a set of Simple[T] providers with a fallback chain
// and an LRU cache. Load is a Keyed[string, T] function keyed by provider name
// — callers pass the provider name (e.g. "fisu") and get the cached or
// fallback-loaded result.
type SimpleFallbackCache[T any] struct {
	fallbacks *containers.Fallbacks[Simple[T]]
	Load      Keyed[string, T]
}

// NewSimpleFallbackCache creates a SimpleFallbackCache.
//
// cacheSize should be len(loaders)+1 to accommodate all provider-specific
// entries plus the fallback result in the LRU.
func NewSimpleFallbackCache[T any](
	log *slog.Logger,
	loaders map[string]Simple[T],
	priority []string,
	cacheSize int,
) *SimpleFallbackCache[T] {
	fallbacks := containers.NewFallbacks(
		log.With("sub", "fallback"),
		loaders, priority, fallbackLastSuccessTTL,
	)
	fallbackLoader := NewFallback(fallbacks)
	cached := WithKeyedCache(
		log.With("sub", "cache"),
		func(ctx context.Context, provider string) (T, error) {
			if ldr, ok := loaders[provider]; ok {
				return ldr(ctx)
			}
			return fallbackLoader(ctx)
		},
		memory.NewKeyedExpirableCache(
			expirable.NewLRU[string, T](
				cacheSize, nil, cacheEntryTTL,
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

// KeyedFallbackCache wraps a set of Keyed[K, T] providers with a fallback
// chain and an LRU cache. Load is a Queried[Query[K], T] function keyed by
// provider name + query key (e.g. provider "fisu" + world ID) — callers pass
// a Query[K] and get the cached or fallback-loaded result.
type KeyedFallbackCache[K comparable, T any] struct {
	fallbacks *containers.Fallbacks[Keyed[K, T]]
	Load      Queried[Query[K], T]
}

// NewKeyedFallbackCache creates a KeyedFallbackCache.
//
// cacheSize should be (len(loaders)+1) * len(possible keys) to accommodate
// all provider/key combinations plus the fallback results in the LRU.
func NewKeyedFallbackCache[K comparable, T any](
	log *slog.Logger,
	loaders map[string]Keyed[K, T],
	priority []string,
	cacheSize int,
) *KeyedFallbackCache[K, T] {
	fallbacks := containers.NewFallbacks(
		log.With("sub", "fallback"),
		loaders, priority, fallbackLastSuccessTTL,
	)
	fallbackLoader := NewKeyedFallback(fallbacks)
	cached := WithQueriedCache(
		log.With("sub", "cache"),
		func(ctx context.Context, query Query[K]) (T, error) {
			if ldr, ok := loaders[query.Provider]; ok {
				return ldr(ctx, query.Key)
			}
			return fallbackLoader(ctx, query.Key)
		},
		memory.NewKeyedExpirableCache(
			expirable.NewLRU[Query[K], T](
				cacheSize, nil, cacheEntryTTL,
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
