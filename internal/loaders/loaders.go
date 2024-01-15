package loaders

import (
	"context"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

var ErrNotFound = fmt.Errorf("not found")

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
	Load(ctx context.Context) (T, error)
}

type QueriedLoader[Q any, T any] interface {
	Load(ctx context.Context, query Q) (T, error)
}

type KeyedLoader[K comparable, T any] QueriedLoader[K, T]

type FallbackLoader[T any] struct {
	fallbacks *containers.Fallbacks[Loader[T]]
}

func NewFallbackLoader[T any](name string, loaders map[string]Loader[T], priority []string) *FallbackLoader[T] {
	return &FallbackLoader[T]{
		fallbacks: containers.NewFallbacks(name, loaders, priority, time.Hour),
	}
}

func (l *FallbackLoader[T]) Start(ctx context.Context) {
	l.fallbacks.Start(ctx)
}

func (l *FallbackLoader[T]) Stop() {
	l.fallbacks.Stop()
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
	name string,
	loaders map[string]KeyedLoader[K, T],
	priority []string,
) *KeyedFallbackLoader[K, T] {
	return &KeyedFallbackLoader[K, T]{
		fallbacks: containers.NewFallbacks(name, loaders, priority, time.Hour),
	}
}

func (l *KeyedFallbackLoader[K, T]) Start(ctx context.Context) {
	l.fallbacks.Start(ctx)
}

func (l *KeyedFallbackLoader[K, T]) Stop() {
	l.fallbacks.Stop()
}

func (l *KeyedFallbackLoader[K, T]) Load(ctx context.Context, key K) (T, error) {
	return containers.ExecFallback(l.fallbacks, func(loader KeyedLoader[K, T]) (T, error) {
		return loader.Load(ctx, key)
	})
}

var ErrLoaderNotFound = fmt.Errorf("loader not found")

type MultiLoader[T any] struct {
	loaders map[string]Loader[T]
}

func NewMultiLoader[T any](loaders map[string]Loader[T]) *MultiLoader[T] {
	return &MultiLoader[T]{loaders}
}

func (l *MultiLoader[T]) Load(ctx context.Context, loader string) (T, error) {
	if loader, ok := l.loaders[loader]; ok {
		return loader.Load(ctx)
	}
	var t T
	return t, fmt.Errorf("unknown loader %q: %w", loader, ErrLoaderNotFound)
}

type KeyedMultiLoader[K comparable, T any] struct {
	loaders map[string]KeyedLoader[K, T]
}

func NewKeyedMultiLoader[K comparable, T any](loaders map[string]KeyedLoader[K, T]) *KeyedMultiLoader[K, T] {
	return &KeyedMultiLoader[K, T]{loaders}
}

type MultiLoaderQuery[K comparable] struct {
	Loader string
	Key    K
}

func NewMultiLoaderQuery[K comparable](loader string, key K) MultiLoaderQuery[K] {
	return MultiLoaderQuery[K]{
		Loader: loader,
		Key:    key,
	}
}

func (l *KeyedMultiLoader[K, T]) Load(ctx context.Context, q MultiLoaderQuery[K]) (T, error) {
	if loader, ok := l.loaders[q.Loader]; ok {
		return loader.Load(ctx, q.Key)
	}
	var t T
	return t, fmt.Errorf("unknown loader %q: %w", q.Loader, ErrLoaderNotFound)
}
