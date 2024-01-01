package ps2

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type Loaded[T any] struct {
	Value     T
	Source    string
	UpdatedAt time.Time
}

func LoadedNow[T any](source string, value T) Loaded[T] {
	return Loaded[T]{
		Value:     value,
		Source:    source,
		UpdatedAt: time.Now(),
	}
}

type Loader[T any] interface {
	Load(ctx context.Context) (Loaded[T], error)
}

type KeyedLoader[K comparable, T any] interface {
	Load(ctx context.Context, key K) (Loaded[T], error)
}

type fallbackLoader[T any] struct {
	fallbacks *containers.Fallbacks[Loader[T]]
}

func NewFallbackLoader[T any](name string, loaders map[string]Loader[T], priority []string) *fallbackLoader[T] {
	return &fallbackLoader[T]{
		fallbacks: containers.NewFallbacks(name, loaders, priority, time.Hour),
	}
}

func (l *fallbackLoader[T]) Start() {
	l.fallbacks.Start()
}

func (l *fallbackLoader[T]) Stop() {
	l.fallbacks.Stop()
}

func (l *fallbackLoader[T]) Load(ctx context.Context) (Loaded[T], error) {
	return containers.ExecFallback(l.fallbacks, func(loader Loader[T]) (Loaded[T], error) {
		return loader.Load(ctx)
	})
}

type keyedFallbackLoader[K comparable, T any] struct {
	fallbacks *containers.Fallbacks[KeyedLoader[K, T]]
}

func NewKeyedFallbackLoader[K comparable, T any](
	name string,
	loaders map[string]KeyedLoader[K, T],
	priority []string,
) *keyedFallbackLoader[K, T] {
	return &keyedFallbackLoader[K, T]{
		fallbacks: containers.NewFallbacks(name, loaders, priority, time.Hour),
	}
}

func (l *keyedFallbackLoader[K, T]) Start() {
	l.fallbacks.Start()
}

func (l *keyedFallbackLoader[K, T]) Stop() {
	l.fallbacks.Stop()
}

func (l *keyedFallbackLoader[K, T]) Load(ctx context.Context, key K) (Loaded[T], error) {
	return containers.ExecFallback(l.fallbacks, func(loader KeyedLoader[K, T]) (Loaded[T], error) {
		return loader.Load(ctx, key)
	})
}
