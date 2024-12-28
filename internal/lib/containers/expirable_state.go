package containers

import (
	"context"
	"sync"
	"time"
)

type ExpirableState[K comparable, S any] struct {
	ttl    time.Duration
	mu     sync.RWMutex
	keys   *ExpirationQueue[K]
	values map[K]S
}

func NewExpirableState[K comparable, S any](ttl time.Duration) *ExpirableState[K, S] {
	return &ExpirableState[K, S]{
		ttl:    ttl,
		keys:   NewExpirationQueue[K](),
		values: make(map[K]S),
	}
}

func (c *ExpirableState[K, S]) flush(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys.RemoveExpired(now.Add(-c.ttl), func(key K) {
		delete(c.values, key)
	})
}

func (c *ExpirableState[K, S]) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			c.flush(now)
		}
	}
}

func (c *ExpirableState[K, S]) Load(key K) (S, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	return value, ok
}

func (c *ExpirableState[K, S]) Store(key K, value S) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys.Push(key)
	c.values[key] = value
}

func (c *ExpirableState[K, S]) Pop(key K) (S, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.values[key]
	if ok {
		c.keys.Remove(key)
		delete(c.values, key)
	}
	return value, ok
}

func (c *ExpirableState[K, S]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys.Remove(key)
	delete(c.values, key)
}
