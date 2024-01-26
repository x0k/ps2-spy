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

func (v *CachedLoader[T]) Cached(ctx context.Context) (T, bool) {
	return v.value.Get(ctx)
}

func (v *CachedLoader[T]) Load(ctx context.Context) (T, error) {
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

type CachedQueryLoader[Q any, T any] struct {
	cache  containers.QueryCache[Q, T]
	loader QueriedLoader[Q, T]
}

func NewCachedQueriedLoader[Q any, T any](
	loader QueriedLoader[Q, T],
	cache containers.QueryCache[Q, T],
) *CachedQueryLoader[Q, T] {
	return &CachedQueryLoader[Q, T]{
		cache:  cache,
		loader: loader,
	}
}

func (v *CachedQueryLoader[Q, T]) Cached(ctx context.Context, query Q) (T, bool) {
	return v.cache.Get(ctx, query)
}

func (v *CachedQueryLoader[Q, T]) Load(ctx context.Context, query Q) (T, error) {
	cached, ok := v.cache.Get(ctx, query)
	if ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx, query)
	if err != nil {
		return loaded, err
	}
	v.cache.Add(ctx, query, loaded)
	return loaded, nil
}

type CachedMultiKeyedLoader[K comparable, T any] struct {
	cache  containers.MultiKeyedCache[K, T]
	loader QueriedLoader[[]K, map[K]T]
}

func NewCtxCachedMultiKeyedLoader[K comparable, T any](
	loader QueriedLoader[[]K, map[K]T],
	cache containers.MultiKeyedCache[K, T],
) *CachedMultiKeyedLoader[K, T] {
	return &CachedMultiKeyedLoader[K, T]{
		cache:  cache,
		loader: loader,
	}
}

// Return true if all keys are cached
func (l *CachedMultiKeyedLoader[K, T]) Cached(ctx context.Context, keys []K) (map[K]T, bool) {
	return l.cache.Get(ctx, keys)
}

func (l *CachedMultiKeyedLoader[K, T]) Load(ctx context.Context, keys []K) (map[K]T, error) {
	if len(keys) == 0 {
		return make(map[K]T), nil
	}
	cached, ok := l.cache.Get(ctx, keys)
	if ok {
		return cached, nil
	}
	toLoad := make([]K, 0, len(keys))
	for _, k := range keys {
		if _, ok := cached[k]; !ok {
			toLoad = append(toLoad, k)
		}
	}
	loaded, err := l.loader.Load(ctx, toLoad)
	if err != nil {
		return cached, err
	}
	l.cache.Add(ctx, loaded)
	for k, v := range cached {
		loaded[k] = v
	}
	return loaded, nil
}
