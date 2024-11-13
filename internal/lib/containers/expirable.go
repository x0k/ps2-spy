package containers

import (
	"context"
	"sync"
	"time"
)

type Expirable[T any] struct {
	mu     sync.RWMutex
	actual bool
	val    T
	ttl    time.Duration
	timer  *time.Timer
}

func NewExpirable[T any](ttl time.Duration) *Expirable[T] {
	return &Expirable[T]{
		ttl: ttl,
	}
}

func (e *Expirable[T]) markAsExpired() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.actual = false
}

func (e *Expirable[T]) Start(ctx context.Context) {
	e.timer = time.NewTimer(e.ttl)
	defer e.timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.timer.C:
			e.markAsExpired()
		}
	}
}

func (e *Expirable[T]) Get() (T, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.val, e.actual
}

func (e *Expirable[T]) Set(val T) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.val = val
	e.actual = true
	if !e.timer.Stop() {
		select {
		case <-e.timer.C:
		default:
		}
	}
	e.timer.Reset(e.ttl)
}
