package ps2

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
)

type Loaded[T any] struct {
	Value     T
	Source    string
	UpdatedAt time.Time
}

type loader[T any] interface {
	Load(ctx context.Context) (T, error)
}

type keyedLoader[K comparable, T any] interface {
	Load(ctx context.Context, key K) (T, error)
}

type loadedLoader[T any] struct {
	name   string
	loader loader[T]
}

func WithLoaded[T any](name string, loader loader[T]) *loadedLoader[T] {
	return &loadedLoader[T]{name, loader}
}

func (l *loadedLoader[T]) Load(ctx context.Context) (Loaded[T], error) {
	value, err := l.loader.Load(ctx)
	if err != nil {
		return Loaded[T]{}, err
	}
	return Loaded[T]{
		Value:     value,
		Source:    l.name,
		UpdatedAt: time.Now(),
	}, nil
}

type keyedLoadedLoader[K comparable, T any] struct {
	name   string
	loader keyedLoader[K, T]
}

func WithKeyedLoaded[K comparable, T any](name string, loader keyedLoader[K, T]) *keyedLoadedLoader[K, T] {
	return &keyedLoadedLoader[K, T]{name, loader}
}

func (l *keyedLoadedLoader[K, T]) Load(ctx context.Context, key K) (Loaded[T], error) {
	value, err := l.loader.Load(ctx, key)
	if err != nil {
		return Loaded[T]{}, err
	}
	return Loaded[T]{
		Value:     value,
		Source:    l.name,
		UpdatedAt: time.Now(),
	}, nil
}

type fallbackLoaderWrapper[T any] struct {
	loaders       []loader[T]
	successLoader *containers.ExpiableValue[loader[T]]
}

func WithFallback[T any](loaders ...loader[T]) *fallbackLoaderWrapper[T] {
	return &fallbackLoaderWrapper[T]{
		loaders:       loaders,
		successLoader: containers.NewExpiableValue[loader[T]](time.Hour),
	}
}

func (l *fallbackLoaderWrapper[T]) Start() {
	go l.successLoader.StartExpiration()
}

func (l *fallbackLoaderWrapper[T]) Stop() {
	l.successLoader.StopExpiration()
}

func (l *fallbackLoaderWrapper[T]) Load(ctx context.Context) (T, error) {
	loader, ok := l.successLoader.Read()
	if ok {
		value, err := loader.Load(ctx)
		if err == nil {
			return value, nil
		}
		l.successLoader.MarkAsExpired()
	}
	for _, loader := range l.loaders {
		value, err := loader.Load(ctx)
		if err != nil {
			continue
		}
		l.successLoader.Write(loader)
		return value, nil
	}
	return *new(T), nil
}

type keyedFallbackLoaderWrapper[K comparable, T any] struct {
	loaders       []keyedLoader[K, T]
	successLoader *containers.ExpiableValue[keyedLoader[K, T]]
}

func WithKeyedFallback[K comparable, T any](loaders ...keyedLoader[K, T]) *keyedFallbackLoaderWrapper[K, T] {
	return &keyedFallbackLoaderWrapper[K, T]{
		loaders:       loaders,
		successLoader: containers.NewExpiableValue[keyedLoader[K, T]](time.Hour),
	}
}

func (l *keyedFallbackLoaderWrapper[K, T]) Start() {
	go l.successLoader.StartExpiration()
}

func (l *keyedFallbackLoaderWrapper[K, T]) Stop() {
	l.successLoader.StopExpiration()
}

func (l *keyedFallbackLoaderWrapper[K, T]) Load(ctx context.Context, key K) (T, error) {
	loader, ok := l.successLoader.Read()
	if ok {
		value, err := loader.Load(ctx, key)
		if err == nil {
			return value, nil
		}
		l.successLoader.MarkAsExpired()
	}
	for _, loader := range l.loaders {
		value, err := loader.Load(ctx, key)
		if err != nil {
			continue
		}
		l.successLoader.Write(loader)
		return value, nil
	}
	return *new(T), nil
}
