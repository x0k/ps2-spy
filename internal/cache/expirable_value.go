package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

type ExpiableValue[T any] struct {
	mu     sync.RWMutex
	actual atomic.Bool
	val    T
	ttl    time.Duration
	ticker *time.Ticker
	done   chan struct{}
}

func (e *ExpiableValue[T]) expiration() {
	for {
		select {
		case <-e.ticker.C:
			e.actual.Store(false)
			e.ticker.Stop()
		case <-e.done:
			return
		}
	}
}

func NewExpiableValue[T any](ttl time.Duration) *ExpiableValue[T] {
	val := &ExpiableValue[T]{
		ttl:    ttl,
		ticker: time.NewTicker(ttl),
		done:   make(chan struct{}),
	}
	go val.expiration()
	return val
}

func (e *ExpiableValue[T]) Stop() {
	close(e.done)
}

func (e *ExpiableValue[T]) read() (T, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.val, e.actual.Load()
}

func (e *ExpiableValue[T]) Load(loader func() (T, error)) (T, error) {
	cached, ok := e.read()
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
	e.actual.Store(true)
	e.ticker.Reset(e.ttl)
	return loaded, nil
}
