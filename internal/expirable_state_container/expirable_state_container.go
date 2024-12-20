package expirable_state_container

import (
	"context"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/containers"
)

type ExpirableStateContainer[K comparable, S any] struct {
	ttl    time.Duration
	mu     sync.RWMutex
	keys   *containers.ExpirationQueue[K]
	values map[K]S
}

func New[K comparable, S any](ttl time.Duration) *ExpirableStateContainer[K, S] {
	return &ExpirableStateContainer[K, S]{
		ttl:    ttl,
		keys:   containers.NewExpirationQueue[K](),
		values: make(map[K]S),
	}
}

func (c *ExpirableStateContainer[K, S]) flush(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys.RemoveExpired(now.Add(-c.ttl), func(key K) {
		delete(c.values, key)
	})
}

func (c *ExpirableStateContainer[K, S]) Start(ctx context.Context) {
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

func (c *ExpirableStateContainer[K, S]) Load(key K) (S, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	return value, ok
}

func (c *ExpirableStateContainer[K, S]) Store(key K, value S) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys.Push(key)
	c.values[key] = value
}

func (c *ExpirableStateContainer[K, S]) Pop(key K) (S, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.values[key]
	if ok {
		c.keys.Remove(key)
		delete(c.values, key)
	}
	return value, ok
}
