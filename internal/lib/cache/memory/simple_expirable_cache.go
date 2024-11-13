package memory

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type SimpleExpirableCache[T any] struct {
	cache *containers.Expirable[T]
}

func NewSimpleExpirable[T any](ttl time.Duration) *SimpleExpirableCache[T] {
	return &SimpleExpirableCache[T]{
		cache: containers.NewExpirable[T](ttl),
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
