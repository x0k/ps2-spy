package loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

func NewFallback[T any](fallbacks *containers.Fallbacks[Simple[T]]) Simple[T] {
	return func(ctx context.Context) (T, error) {
		return containers.ExecFallback(fallbacks, func(loader Simple[T]) (T, error) {
			return loader(ctx)
		})
	}
}

func NewKeyedFallback[K comparable, T any](fallbacks *containers.Fallbacks[Keyed[K, T]]) Keyed[K, T] {
	return func(ctx context.Context, key K) (T, error) {
		return containers.ExecFallback(fallbacks, func(loader Keyed[K, T]) (T, error) {
			return loader(ctx, key)
		})
	}
}
