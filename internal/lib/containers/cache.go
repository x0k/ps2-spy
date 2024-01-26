package containers

import (
	"context"
)

type Cache[T any] interface {
	Get() (T, bool)
	Set(value T)
}

type KeyedCache[K comparable, T any] interface {
	Get(key K) (T, bool)
	// Returns bool to compatibility with expirable.LRU
	// eviction indicator
	Add(key K, value T) bool
}

type ContextCache[T any] interface {
	Get(ctx context.Context) (T, bool)
	Set(ctx context.Context, value T)
}

type ContextQueryCache[K any, T any] interface {
	Get(ctx context.Context, query K) (T, bool)
	Add(ctx context.Context, query K, value T) error
}

type ContextKeyedCache[K comparable, T any] ContextQueryCache[K, T]

type ContextMultiKeyedCache[K comparable, T any] interface {
	Get(ctx context.Context, keys []K) (map[K]T, bool)
	Add(ctx context.Context, values map[K]T) error
}
