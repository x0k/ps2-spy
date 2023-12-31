package ps2

import (
	"context"
	"log"
	"time"

	"github.com/x0k/ps2-spy/internal/containers"
)

type Loaded[T any] struct {
	Value     T
	Source    string
	UpdatedAt time.Time
}

type loader[T any] interface {
	Name() string
	Load(ctx context.Context) (T, error)
}

type keyedLoader[K comparable, T any] interface {
	Name() string
	Load(ctx context.Context, key K) (T, error)
}

type loadedLoader[T any] struct {
	loader loader[T]
}

func WithLoaded[T any](loader loader[T]) *loadedLoader[T] {
	return &loadedLoader[T]{loader}
}

func (l *loadedLoader[T]) Name() string {
	return l.loader.Name()
}

func (l *loadedLoader[T]) Load(ctx context.Context) (Loaded[T], error) {
	value, err := l.loader.Load(ctx)
	if err != nil {
		return Loaded[T]{}, err
	}
	return Loaded[T]{
		Value:     value,
		Source:    l.loader.Name(),
		UpdatedAt: time.Now(),
	}, nil
}

type keyedLoadedLoader[K comparable, T any] struct {
	loader keyedLoader[K, T]
}

func WithKeyedLoaded[K comparable, T any](loader keyedLoader[K, T]) *keyedLoadedLoader[K, T] {
	return &keyedLoadedLoader[K, T]{loader}
}

func (l *keyedLoadedLoader[K, T]) Name() string {
	return l.loader.Name()
}

func (l *keyedLoadedLoader[K, T]) Load(ctx context.Context, key K) (Loaded[T], error) {
	value, err := l.loader.Load(ctx, key)
	if err != nil {
		return Loaded[T]{}, err
	}
	return Loaded[T]{
		Value:     value,
		Source:    l.loader.Name(),
		UpdatedAt: time.Now(),
	}, nil
}

type fallbackLoader[T any] struct {
	name          string
	loaders       []loader[T]
	successLoader *containers.ExpiableValue[loader[T]]
}

func WithFallback[T any](name string, loaders ...loader[T]) *fallbackLoader[T] {
	return &fallbackLoader[T]{
		name:          name,
		loaders:       loaders,
		successLoader: containers.NewExpiableValue[loader[T]](time.Hour),
	}
}

func (l *fallbackLoader[T]) Name() string {
	if loader, ok := l.successLoader.Read(); ok {
		return loader.Name()
	}
	return l.name
}

func (l *fallbackLoader[T]) Start() {
	go l.successLoader.StartExpiration()
}

func (l *fallbackLoader[T]) Stop() {
	l.successLoader.StopExpiration()
}

func (l *fallbackLoader[T]) Load(ctx context.Context) (T, error) {
	loader, ok := l.successLoader.Read()
	if ok {
		value, err := loader.Load(ctx)
		if err == nil {
			return value, nil
		}
		log.Printf("[%s] Last successful loader %q failed: %q", l.name, loader.Name(), err)
		l.successLoader.MarkAsExpired()
	}
	for _, loader := range l.loaders {
		value, err := loader.Load(ctx)
		if err != nil {
			log.Printf("[%s] Loader %q failed: %q", l.name, loader.Name(), err)
			continue
		}
		l.successLoader.Write(loader)
		return value, nil
	}
	return *new(T), nil
}

type keyedFallbackLoader[K comparable, T any] struct {
	name          string
	loaders       []keyedLoader[K, T]
	successLoader *containers.ExpiableValue[keyedLoader[K, T]]
}

func WithKeyedFallback[K comparable, T any](name string, loaders ...keyedLoader[K, T]) *keyedFallbackLoader[K, T] {
	return &keyedFallbackLoader[K, T]{
		name:          name,
		loaders:       loaders,
		successLoader: containers.NewExpiableValue[keyedLoader[K, T]](time.Hour),
	}
}

func (l *keyedFallbackLoader[K, T]) Name() string {
	if loader, ok := l.successLoader.Read(); ok {
		return loader.Name()
	}
	return l.name
}

func (l *keyedFallbackLoader[K, T]) Start() {
	go l.successLoader.StartExpiration()
}

func (l *keyedFallbackLoader[K, T]) Stop() {
	l.successLoader.StopExpiration()
}

func (l *keyedFallbackLoader[K, T]) Load(ctx context.Context, key K) (T, error) {
	loader, ok := l.successLoader.Read()
	if ok {
		value, err := loader.Load(ctx, key)
		if err == nil {
			return value, nil
		}
		log.Printf("[%s] Last successful loader %q failed: %q", l.name, loader.Name(), err)
		l.successLoader.MarkAsExpired()
	}
	for _, loader := range l.loaders {
		value, err := loader.Load(ctx, key)
		if err != nil {
			log.Printf("[%s] Loader %q failed: %q", l.name, loader.Name(), err)
			continue
		}
		l.successLoader.Write(loader)
		return value, nil
	}
	return *new(T), nil
}
