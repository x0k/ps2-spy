package cache

import (
	"context"
)

type Simple[T any] interface {
	Get(ctx context.Context) (T, bool)
	Add(ctx context.Context, value T) error
}

type Queried[K any, T any] interface {
	Get(ctx context.Context, query K) (T, bool)
	Add(ctx context.Context, query K, value T) error
}

type Keyed[K comparable, T any] Queried[K, T]

type Multi[K comparable, T any] interface {
	Get(ctx context.Context, keys []K) (map[K]T, bool)
	Add(ctx context.Context, values map[K]T) error
}
