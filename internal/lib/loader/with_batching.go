package loader

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Batched[K comparable, T any] struct {
	loader        Multi[K, T]
	awaitersMu    sync.Mutex
	awaiters      map[K][]chan result[T]
	checkRate     time.Duration
	notFoundError error
}

type result[T any] struct {
	value T
	err   error
}

func WithBatching[K comparable, T any](
	loader Multi[K, T],
	checkRate time.Duration,
	notFoundError error,
) *Batched[K, T] {
	return &Batched[K, T]{
		loader:        loader,
		awaiters:      make(map[K][]chan result[T]),
		checkRate:     checkRate,
		notFoundError: notFoundError,
	}
}

func (b *Batched[K, T]) batch() []K {
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	batch := make([]K, 0, len(b.awaiters))
	for key := range b.awaiters {
		batch = append(batch, key)
	}
	return batch
}

func (b *Batched[K, T]) releaseAwaiters(_ context.Context, batch []K, results map[K]T) {
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	for _, key := range batch {
		if channels, ok := b.awaiters[key]; ok {
			var res result[T]
			if val, ok := results[key]; ok {
				res.value = val
			} else {
				res.err = fmt.Errorf("awaiter result not found: %w", b.notFoundError)
			}
			for _, channel := range channels {
				channel <- res
				close(channel)
			}
			delete(b.awaiters, key)
		}
	}
}

func (b *Batched[K, T]) processNonEmptyBatch(ctx context.Context, batch []K) {
	results, err := b.loader(ctx, batch)
	if err != nil {
		results = make(map[K]T)
	}
	b.releaseAwaiters(ctx, batch, results)
}

func (b *Batched[K, T]) Start(ctx context.Context) {
	ticker := time.NewTicker(b.checkRate)
	defer ticker.Stop()
	alive := true
	for alive {
		select {
		case <-ctx.Done():
			// break is not propagated to for loop ?
			alive = false
		case <-ticker.C:
			batch := b.batch()
			if len(batch) == 0 {
				continue
			}
			b.processNonEmptyBatch(ctx, batch)
		}
	}
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	for _, channels := range b.awaiters {
		for _, c := range channels {
			close(c)
		}
	}
	clear(b.awaiters)
}

func (b *Batched[K, T]) load(key K) chan result[T] {
	c := make(chan result[T])
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	b.awaiters[key] = append(b.awaiters[key], c)
	return c
}

func (b *Batched[K, T]) Load(ctx context.Context, key K) (T, error) {
	select {
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	case r := <-b.load(key):
		return r.value, r.err
	}
}
