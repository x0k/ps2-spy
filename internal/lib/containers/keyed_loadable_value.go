package containers

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type keyedValueLoader[K comparable, T any] interface {
	Load(ctx context.Context, key K) (T, error)
}

func NewKeyedLoadableValue[K comparable, T any](
	loader keyedValueLoader[K, T],
	maxSize int,
	ttl time.Duration,
) *QueriedLoadableValue[K, K, T] {
	return &QueriedLoadableValue[K, K, T]{
		cache:  expirable.NewLRU[K, T](maxSize, nil, ttl),
		loader: loader,
		mapper: func(k K) K { return k },
	}
}
