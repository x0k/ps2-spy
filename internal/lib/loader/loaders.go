package loader

import "context"

type Simple[T any] func(context.Context) (T, error)

type Queried[Q any, T any] func(context.Context, Q) (T, error)

type Keyed[K comparable, T any] Queried[K, T]

type Multi[K comparable, T any] func(context.Context, []K) (map[K]T, error)
