package ps2

import (
	"context"
	"fmt"
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

type QueriedLoader[Q any, T any] interface {
	Load(ctx context.Context, query Q) (Loaded[T], error)
}

type KeyedLoader[K comparable, T any] QueriedLoader[K, T]

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

var ErrLoaderNotFound = fmt.Errorf("loader not found")

type multiLoader[T any] struct {
	loaders map[string]Loader[T]
}

func NewMultiLoader[T any](loaders map[string]Loader[T]) *multiLoader[T] {
	return &multiLoader[T]{loaders}
}

func (l *multiLoader[T]) Load(ctx context.Context, loader string) (Loaded[T], error) {
	if loader, ok := l.loaders[loader]; ok {
		return loader.Load(ctx)
	}
	return Loaded[T]{}, fmt.Errorf("unknown loader %q: %w", loader, ErrLoaderNotFound)
}

type keyedMultiLoader[K comparable, T any] struct {
	loaders map[string]KeyedLoader[K, T]
}

func NewKeyedMultiLoader[K comparable, T any](loaders map[string]KeyedLoader[K, T]) *keyedMultiLoader[K, T] {
	return &keyedMultiLoader[K, T]{loaders}
}

type multiLoaderQuery[K comparable] struct {
	loader string
	key    K
}

func (l *keyedMultiLoader[K, T]) Load(ctx context.Context, q multiLoaderQuery[K]) (Loaded[T], error) {
	if loader, ok := l.loaders[q.loader]; ok {
		return loader.Load(ctx, q.key)
	}
	return Loaded[T]{}, fmt.Errorf("unknown loader %q: %w", q.loader, ErrLoaderNotFound)
}
