package memory

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type SimpleExpirableCache[T any] struct {
	cache *containers.Expirable[T]
}

func NewSimpleExpirable[T any](cache *containers.Expirable[T]) *SimpleExpirableCache[T] {
	return &SimpleExpirableCache[T]{
		cache: cache,
	}
}

func (c *SimpleExpirableCache[T]) Start(ctx context.Context) {
	c.cache.Start(ctx)
}

func (c *SimpleExpirableCache[T]) Get(ctx context.Context) (T, bool) {
	return c.cache.Get()
}

func (c *SimpleExpirableCache[T]) Add(ctx context.Context, value T) error {
	c.cache.Set(value)
	return nil
}
