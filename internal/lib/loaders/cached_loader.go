package loaders

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type CachedLoader[T any] struct {
	value  containers.Cache[T]
	loader Loader[T]
}

func NewCachedLoader[T any](loader Loader[T], cache containers.Cache[T]) *CachedLoader[T] {
	return &CachedLoader[T]{
		value:  cache,
		loader: loader,
	}
}

func (v *CachedLoader[T]) Cached() (T, bool) {
	return v.value.Get()
}

func (v *CachedLoader[T]) Load(ctx context.Context) (T, error) {
	if cached, ok := v.value.Get(); ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx)
	if err != nil {
		return loaded, err
	}
	v.value.Set(loaded)
	return loaded, nil
}

type CachedQueryLoader[Q any, K comparable, T any] struct {
	cache  containers.KeyedCache[K, T]
	loader QueriedLoader[Q, T]
	mapper func(Q) K
}

func NewCachedQueriedLoader[Q any, K comparable, T any](
	loader QueriedLoader[Q, T],
	cache containers.KeyedCache[K, T],
	mapper func(Q) K,
) *CachedQueryLoader[Q, K, T] {
	return &CachedQueryLoader[Q, K, T]{
		cache:  cache,
		loader: loader,
		mapper: mapper,
	}
}

func (v *CachedQueryLoader[Q, K, T]) Cached(query Q) (T, bool) {
	key := v.mapper(query)
	return v.cache.Get(key)
}

func (v *CachedQueryLoader[Q, K, T]) Load(ctx context.Context, query Q) (T, error) {
	key := v.mapper(query)
	cached, ok := v.cache.Get(key)
	if ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx, query)
	if err != nil {
		return loaded, err
	}
	v.cache.Add(key, loaded)
	return loaded, nil
}

func NewCachedKeyedLoader[K comparable, T any](
	loader KeyedLoader[K, T],
	cache containers.KeyedCache[K, T],
) *CachedQueryLoader[K, K, T] {
	return &CachedQueryLoader[K, K, T]{
		cache:  cache,
		loader: loader,
		mapper: func(k K) K { return k },
	}
}
