package containers

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type Expiable[T any] struct {
	mu      sync.RWMutex
	actual  atomic.Bool
	val     T
	ttl     time.Duration
	ticker  *time.Ticker
	started atomic.Bool
}

func NewExpiable[T any](ttl time.Duration) *Expiable[T] {
	return &Expiable[T]{
		ttl: ttl,
	}
}

func (e *Expiable[T]) MarkAsExpired() {
	e.actual.Store(false)
	e.ticker.Stop()
}

func (e *Expiable[T]) ResetExpiration() {
	e.actual.Store(true)
	e.ticker.Reset(e.ttl)
}

func (e *Expiable[T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	if e.started.Swap(true) {
		return
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.ticker = time.NewTicker(e.ttl)
		defer e.ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-e.ticker.C:
				e.MarkAsExpired()
			}
		}
	}()
}

func (e *Expiable[T]) Get() (T, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.val, e.actual.Load()
}

func (e *Expiable[T]) Set(val T) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.val = val
	e.ResetExpiration()
}

func (e *Expiable[T]) Load(loader func() (T, error)) (T, error) {
	cached, ok := e.Get()
	if ok {
		return cached, nil
	}
	loaded, err := loader()
	if err != nil {
		return cached, err
	}
	e.Set(loaded)
	return loaded, nil
}

type ExpiableLRU[K comparable, T any] struct {
	cache *expirable.LRU[K, T]
}

func NewExpiableLRU[K comparable, T any](size int, ttl time.Duration) *ExpiableLRU[K, T] {
	return &ExpiableLRU[K, T]{
		cache: expirable.NewLRU[K, T](size, nil, ttl),
	}
}

func (e *ExpiableLRU[K, T]) Get(_ context.Context, key K) (T, bool) {
	return e.cache.Get(key)
}

func (e *ExpiableLRU[K, T]) Add(_ context.Context, key K, value T) error {
	e.cache.Add(key, value)
	return nil
}
