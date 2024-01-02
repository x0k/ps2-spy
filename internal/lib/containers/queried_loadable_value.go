package containers

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type queriedValueLoader[Q any, T any] interface {
	Load(ctx context.Context, query Q) (T, error)
}

type QueriedLoadableValue[Q any, K comparable, T any] struct {
	cache  *expirable.LRU[K, T]
	loader queriedValueLoader[Q, T]
	mapper func(Q) K
}

func NewQueriedLoadableValue[Q any, K comparable, T any](
	loader queriedValueLoader[Q, T],
	maxSize int,
	ttl time.Duration,
	mapper func(Q) K,
) *QueriedLoadableValue[Q, K, T] {
	return &QueriedLoadableValue[Q, K, T]{
		cache:  expirable.NewLRU[K, T](maxSize, nil, ttl),
		loader: loader,
		mapper: mapper,
	}
}

func (v *QueriedLoadableValue[Q, K, T]) Load(ctx context.Context, query Q) (T, error) {
	key := v.mapper(query)
	cached, ok := v.cache.Get(key)
	if ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx, query)
	if err != nil {
		return loaded, err
	}
	v.cache.Add(key, loaded)
	return loaded, nil
}
