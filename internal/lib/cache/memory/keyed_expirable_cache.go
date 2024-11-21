package memory

import (
	"context"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type KeyedExpirableCache[K comparable, T any] struct {
	cache *expirable.LRU[K, T]
}

func NewKeyedExpirableCache[K comparable, T any](cache *expirable.LRU[K, T]) *KeyedExpirableCache[K, T] {
	return &KeyedExpirableCache[K, T]{
		cache: cache,
	}
}

func (c *KeyedExpirableCache[K, T]) Get(_ context.Context, query K) (T, bool) {
	return c.cache.Get(query)
}

func (c *KeyedExpirableCache[K, T]) Add(_ context.Context, query K, value T) error {
	c.cache.Add(query, value)
	return nil
}
