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

type CtxCachedQueryLoader[Q any, T any] struct {
	cache  containers.ContextQueryCache[Q, T]
	loader QueriedLoader[Q, T]
}

func NewCtxCachedQueriedLoader[Q any, T any](
	loader QueriedLoader[Q, T],
	cache containers.ContextQueryCache[Q, T],
) *CtxCachedQueryLoader[Q, T] {
	return &CtxCachedQueryLoader[Q, T]{
		cache:  cache,
		loader: loader,
	}
}

func (v *CtxCachedQueryLoader[Q, T]) Cached(ctx context.Context, query Q) (T, bool) {
	return v.cache.Get(ctx, query)
}

func (v *CtxCachedQueryLoader[Q, T]) Load(ctx context.Context, query Q) (T, error) {
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

type CtxCachedMultiKeyedLoader[K comparable, T any] struct {
	cache  containers.ContextMultiKeyedCache[K, T]
	loader QueriedLoader[[]K, map[K]T]
}

func NewCtxCachedMultiKeyedLoader[K comparable, T any](
	loader QueriedLoader[[]K, map[K]T],
	cache containers.ContextMultiKeyedCache[K, T],
) *CtxCachedMultiKeyedLoader[K, T] {
	return &CtxCachedMultiKeyedLoader[K, T]{
		cache:  cache,
		loader: loader,
	}
}

// Return true if all keys are cached
func (l *CtxCachedMultiKeyedLoader[K, T]) Cached(ctx context.Context, keys []K) (map[K]T, bool) {
	return l.cache.Get(ctx, keys)
}

func (l *CtxCachedMultiKeyedLoader[K, T]) Load(ctx context.Context, keys []K) (map[K]T, error) {
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
