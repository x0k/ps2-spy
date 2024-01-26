package containers

import (
	"context"
)

type Cache[T any] interface {
	Get(ctx context.Context) (T, bool)
	Set(ctx context.Context, value T)
}

type QueryCache[K any, T any] interface {
	Get(ctx context.Context, query K) (T, bool)
	Add(ctx context.Context, query K, value T) error
}

type KeyedCache[K comparable, T any] QueryCache[K, T]

type MultiKeyedCache[K comparable, T any] interface {
	Get(ctx context.Context, keys []K) (map[K]T, bool)
	Add(ctx context.Context, values map[K]T) error
}
