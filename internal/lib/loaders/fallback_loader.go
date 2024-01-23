package loaders

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type FallbackLoader[T any] struct {
	fallbacks *containers.Fallbacks[Loader[T]]
}

func NewFallbackLoader[T any](log *slog.Logger, loaders map[string]Loader[T], priority []string) *FallbackLoader[T] {
	return &FallbackLoader[T]{
		fallbacks: containers.NewFallbacks(log, loaders, priority, time.Hour),
	}
}

func (l *FallbackLoader[T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	l.fallbacks.Start(ctx, wg)
}

func (l *FallbackLoader[T]) Load(ctx context.Context) (T, error) {
	return containers.ExecFallback(l.fallbacks, func(loader Loader[T]) (T, error) {
		return loader.Load(ctx)
	})
}

type KeyedFallbackLoader[K comparable, T any] struct {
	fallbacks *containers.Fallbacks[KeyedLoader[K, T]]
}

func NewKeyedFallbackLoader[K comparable, T any](
	log *slog.Logger,
	loaders map[string]KeyedLoader[K, T],
	priority []string,
) *KeyedFallbackLoader[K, T] {
	return &KeyedFallbackLoader[K, T]{
		fallbacks: containers.NewFallbacks(log, loaders, priority, time.Hour),
	}
}

func (l *KeyedFallbackLoader[K, T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	l.fallbacks.Start(ctx, wg)
}

func (l *KeyedFallbackLoader[K, T]) Load(ctx context.Context, key K) (T, error) {
	return containers.ExecFallback(l.fallbacks, func(loader KeyedLoader[K, T]) (T, error) {
		return loader.Load(ctx, key)
	})
}
