package loaders

import (
	"context"
	"fmt"
)

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
