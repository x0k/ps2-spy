package ps2

import (
	"context"
	"time"
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

type loaderWrapper[T any] struct {
	name   string
	loader loader[T]
}

func WithLoaded[T any](name string, loader loader[T]) loader[Loaded[T]] {
	return &loaderWrapper[T]{name, loader}
}

func (l *loaderWrapper[T]) Load(ctx context.Context) (Loaded[T], error) {
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

type keyedLoaderWrapper[K comparable, T any] struct {
	name   string
	loader keyedLoader[K, T]
}

func WithKeyedLoaded[K comparable, T any](name string, loader keyedLoader[K, T]) keyedLoader[K, Loaded[T]] {
	return &keyedLoaderWrapper[K, T]{name, loader}
}

func (l *keyedLoaderWrapper[K, T]) Load(ctx context.Context, key K) (Loaded[T], error) {
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
