package containers

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
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
