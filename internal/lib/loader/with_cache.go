package loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/cache"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

func WithCache[T any](log *logger.Logger, loader Simple[T], cache cache.Simple[T]) Simple[T] {
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
			log.Error(ctx, "failed to add to cache", sl.Err(err))
		}
		return loaded, nil
	}
}

func WithQueriedCache[K any, T any](
	log *logger.Logger,
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
			log.Error(ctx, "failed to add to cache", sl.Err(err))
		}
		return loaded, nil
	}
}
