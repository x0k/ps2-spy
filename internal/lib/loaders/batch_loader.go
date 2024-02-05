package loaders

import (
	"context"
	"sync"
	"time"
)

type BatchLoader[K comparable, T any] struct {
	loader     QueriedLoader[[]K, map[K]T]
	awaitersMu sync.Mutex
	awaiters   map[K][]chan result[T]
	checkRate  time.Duration
}

type result[T any] struct {
	value T
	err   error
}

func NewBatchLoader[K comparable, T any](
	loader QueriedLoader[[]K, map[K]T],
	checkRate time.Duration,
) *BatchLoader[K, T] {
	return &BatchLoader[K, T]{
		loader:    loader,
		awaiters:  make(map[K][]chan result[T]),
		checkRate: checkRate,
	}
}

func (b *BatchLoader[K, T]) batch() []K {
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	batch := make([]K, 0, len(b.awaiters))
	for key := range b.awaiters {
		batch = append(batch, key)
	}
	return batch
}

func (b *BatchLoader[K, T]) releaseAwaiters(ctx context.Context, batch []K, results map[K]T) {
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	for _, key := range batch {
		if channels, ok := b.awaiters[key]; ok {
			var res result[T]
			if val, ok := results[key]; ok {
				res.value = val
			} else {
				res.err = ErrNotFound
			}
			for _, channel := range channels {
				channel <- res
				close(channel)
			}
			delete(b.awaiters, key)
		}
	}
}

func (b *BatchLoader[K, T]) processNonEmptyBatch(ctx context.Context, batch []K) {
	results, err := b.loader.Load(ctx, batch)
	if err != nil {
		results = make(map[K]T)
	}
	b.releaseAwaiters(ctx, batch, results)
}

func (b *BatchLoader[K, T]) processBatchTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(b.checkRate)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch := b.batch()
			if len(batch) == 0 {
				continue
			}
			b.processNonEmptyBatch(ctx, batch)
		}
	}
}

func (b *BatchLoader[K, T]) cleanupTask(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	for _, channels := range b.awaiters {
		for _, c := range channels {
			close(c)
		}
	}
	clear(b.awaiters)
}

func (b *BatchLoader[K, T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(2)
	go b.processBatchTask(ctx, wg)
	go b.cleanupTask(ctx, wg)
}

func (b *BatchLoader[K, T]) load(key K) chan result[T] {
	c := make(chan result[T])
	b.awaitersMu.Lock()
	defer b.awaitersMu.Unlock()
	b.awaiters[key] = append(b.awaiters[key], c)
	return c
}

func (b *BatchLoader[K, T]) Load(ctx context.Context, key K) (T, error) {
	select {
	case <-ctx.Done():
		var t T
		return t, ctx.Err()
	case r := <-b.load(key):
		return r.value, r.err
	}
}
