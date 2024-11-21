package loader

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/cache"
)

func WithCache[T any](
	log *slog.Logger,
	loader Simple[T],
	cache cache.Simple[T],
) Simple[T] {
	return func(ctx context.Context) (T, error) {
		cached, ok := cache.Get(ctx)
		if ok {
			return cached, nil
		}
		loaded, err := loader(ctx)
		if err != nil {
			return loaded, err
		}
		if err := cache.Add(ctx, loaded); err != nil {
			log.LogAttrs(
				ctx,
				slog.LevelError,
				"failed to add to cache",
				slog.String("error", err.Error()),
			)
		}
		return loaded, nil
	}
}

func WithQueriedCache[K any, T any](
	log *slog.Logger,
	loader Queried[K, T],
	cache cache.Queried[K, T],
) Queried[K, T] {
	return func(ctx context.Context, query K) (T, error) {
		cached, ok := cache.Get(ctx, query)
		if ok {
			return cached, nil
		}
		loaded, err := loader(ctx, query)
		if err != nil {
			return loaded, err
		}
		if err := cache.Add(ctx, query, loaded); err != nil {
			log.LogAttrs(
				ctx,
				slog.LevelError,
				"failed to cache loader result",
				slog.String("error", err.Error()),
			)
		}
		return loaded, nil
	}
}

func WithKeyedCache[K comparable, T any](
	log *slog.Logger,
	loader Keyed[K, T],
	cache cache.Keyed[K, T],
) Keyed[K, T] {
	return Keyed[K, T](WithQueriedCache(log, Queried[K, T](loader), cache))
}

func WithMultiCache[K comparable, T any](
	log *slog.Logger,
	loader Multi[K, T],
	cache cache.Multi[K, T],
) Multi[K, T] {
	return func(ctx context.Context, keys []K) (map[K]T, error) {
		if len(keys) == 0 {
			return make(map[K]T), nil
		}
		cached, ok := cache.Get(ctx, keys)
		if ok {
			return cached, nil
		}
		toLoad := make([]K, 0, len(keys))
		for _, k := range keys {
			if _, ok := cached[k]; !ok {
				toLoad = append(toLoad, k)
			}
		}
		loaded, err := loader(ctx, toLoad)
		if err != nil {
			return cached, err
		}
		if err := cache.Add(ctx, loaded); err != nil {
			log.LogAttrs(
				ctx,
				slog.LevelError,
				"failed to cache loader result",
				slog.String("error", err.Error()),
			)
		}
		for k, v := range cached {
			loaded[k] = v
		}
		return loaded, nil
	}
}
