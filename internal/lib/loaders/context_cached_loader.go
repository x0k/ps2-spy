package loaders

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type CtxCachedLoader[T any] struct {
	value  containers.ContextCache[T]
	loader Loader[T]
}

func NewCtxCachedLoader[T any](loader Loader[T], cache containers.ContextCache[T]) *CtxCachedLoader[T] {
	return &CtxCachedLoader[T]{
		value:  cache,
		loader: loader,
	}
}

func (v *CtxCachedLoader[T]) Cached(ctx context.Context) (T, bool) {
	return v.value.Get(ctx)
}

func (v *CtxCachedLoader[T]) Load(ctx context.Context) (T, error) {
	if cached, ok := v.value.Get(ctx); ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx)
	if err != nil {
		return loaded, err
	}
	v.value.Set(ctx, loaded)
	return loaded, nil
}

type CtxCachedQueryLoader[Q any, K comparable, T any] struct {
	cache  containers.ContextKeyedCache[K, T]
	loader QueriedLoader[Q, T]
	mapper func(Q) K
}

func NewCtxQueriedLoader[Q any, K comparable, T any](
	loader QueriedLoader[Q, T],
	cache containers.ContextKeyedCache[K, T],
	mapper func(Q) K,
) *CtxCachedQueryLoader[Q, K, T] {
	return &CtxCachedQueryLoader[Q, K, T]{
		cache:  cache,
		loader: loader,
		mapper: mapper,
	}
}

func (v *CtxCachedQueryLoader[Q, K, T]) Cached(ctx context.Context, query Q) (T, bool) {
	key := v.mapper(query)
	return v.cache.Get(ctx, key)
}

func (v *CtxCachedQueryLoader[Q, K, T]) Load(ctx context.Context, query Q) (T, error) {
	key := v.mapper(query)
	cached, ok := v.cache.Get(ctx, key)
	if ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx, query)
	if err != nil {
		return loaded, err
	}
	v.cache.Add(ctx, key, loaded)
	return loaded, nil
}

func NewCtxCachedKeyedLoader[K comparable, T any](
	loader KeyedLoader[K, T],
	cache containers.ContextKeyedCache[K, T],
) *CtxCachedQueryLoader[K, K, T] {
	return &CtxCachedQueryLoader[K, K, T]{
		cache:  cache,
		loader: loader,
		mapper: func(k K) K { return k },
	}
}
