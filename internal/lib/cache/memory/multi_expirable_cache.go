package memory

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type MultiExpirableCache[K comparable, T any] struct {
	cache *expirable.LRU[K, T]
}

// NOTE: Start a new goroutine for evicting expired items which will never be stopped.
func NewMultiExpirableCache[K comparable, T any](size int, ttl time.Duration) *MultiExpirableCache[K, T] {
	return &MultiExpirableCache[K, T]{
		cache: expirable.NewLRU[K, T](size, nil, ttl),
	}
}

func (c *MultiExpirableCache[K, T]) Get(_ context.Context, keys []K) (map[K]T, bool) {
	result := make(map[K]T, len(keys))
	for _, key := range keys {
		v, ok := c.cache.Get(key)
		if ok {
			result[key] = v
		}
	}
	return result, len(result) == len(keys)
}

func (c *MultiExpirableCache[K, T]) Add(_ context.Context, values map[K]T) error {
	for key, value := range values {
		c.cache.Add(key, value)
	}
	return nil
}
