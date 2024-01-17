package containers

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type ExpiableValue[T any] struct {
	mu      sync.RWMutex
	actual  atomic.Bool
	val     T
	ttl     time.Duration
	ticker  *time.Ticker
	started atomic.Bool
}

func NewExpiableValue[T any](ttl time.Duration) *ExpiableValue[T] {
	return &ExpiableValue[T]{
		ttl:    ttl,
		ticker: time.NewTicker(ttl),
	}
}

func (e *ExpiableValue[T]) MarkAsExpired() {
	e.actual.Store(false)
	e.ticker.Stop()
}

func (e *ExpiableValue[T]) ResetExpiration() {
	e.actual.Store(true)
	e.ticker.Reset(e.ttl)
}

func (e *ExpiableValue[T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	if e.started.Swap(true) {
		return
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
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

func (e *ExpiableValue[T]) Read() (T, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.val, e.actual.Load()
}

func (e *ExpiableValue[T]) Write(val T) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.val = val
	e.ResetExpiration()
}

func (e *ExpiableValue[T]) Load(loader func() (T, error)) (T, error) {
	cached, ok := e.Read()
	if ok {
		return cached, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	loaded, err := loader()
	if err != nil {
		return cached, err
	}
	e.val = loaded
	e.ResetExpiration()
	return loaded, nil
}
