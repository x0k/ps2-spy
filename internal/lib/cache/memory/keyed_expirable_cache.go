package memory

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type KeyedExpirableCache[K comparable, T any] struct {
	cache *expirable.LRU[K, T]
}

// NOTE: Start a new goroutine for evicting expired items which will never be stopped.
func NewKeyedExpirableCache[K comparable, T any](size int, ttl time.Duration) *KeyedExpirableCache[K, T] {
	return &KeyedExpirableCache[K, T]{
		cache: expirable.NewLRU[K, T](size, nil, ttl),
	}
}

func (c *KeyedExpirableCache[K, T]) Get(_ context.Context, query K) (T, bool) {
	return c.cache.Get(query)
}

func (c *KeyedExpirableCache[K, T]) Add(_ context.Context, query K, value T) error {
	c.cache.Add(query, value)
	return nil
}
