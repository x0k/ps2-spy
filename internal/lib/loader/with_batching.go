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
	awaiters      map[K][]chan T
	checkRate     time.Duration
	notFoundError error
}

func WithBatching[K comparable, T any](
	loader Multi[K, T],
	checkRate time.Duration,
	notFoundError error,
) *Batched[K, T] {
	return &Batched[K, T]{
		loader:        loader,
		awaiters:      make(map[K][]chan T),
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
			for _, channel := range channels {
				if val, ok := results[key]; ok {
					channel <- val
				}
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

func (b *Batched[K, T]) load(key K) chan T {
	c := make(chan T)
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	b.awaiters[key] = append(b.awaiters[key], c)
	return c
}

func (b *Batched[K, T]) Load(ctx context.Context, key K) (T, error) {
	select {
	case <-ctx.Done():
		var t T
		return t, fmt.Errorf("failed to load %v: %w", key, ctx.Err())
	case r, ok := <-b.load(key):
		var err error
		if !ok {
			err = fmt.Errorf("failed to load %v: %w", key, b.notFoundError)
		}
		return r, err
	}
}
