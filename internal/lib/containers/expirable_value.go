package containers

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

func (e *ExpiableValue[T]) isStarted() bool {
	if e.done == nil {
		return false
	}
	select {
	case <-e.done:
		return false
	default:
		return true
	}
}

func (e *ExpiableValue[T]) StartExpiration() {
	if e.isStarted() {
		return
	}
	e.done = make(chan struct{})
	defer e.ticker.Stop()
	for {
		select {
		case <-e.ticker.C:
			e.MarkAsExpired()
		case <-e.done:
			return
		}
	}
}

func (e *ExpiableValue[T]) StopExpiration() {
	close(e.done)
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
