package containers

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

type keyedValueLoader[K comparable, T any] interface {
	Load(ctx context.Context, key K) (T, error)
}

type KeyedLoadableValues[K comparable, T any] struct {
	cache  *expirable.LRU[K, T]
	loader keyedValueLoader[K, T]
}

func NewKeyedLoadableValue[K comparable, T any](
	loader keyedValueLoader[K, T],
	maxSize int,
	ttl time.Duration,
) *KeyedLoadableValues[K, T] {
	return &KeyedLoadableValues[K, T]{
		cache: expirable.NewLRU[K, T](maxSize, nil, ttl),
	}
}

func (v *KeyedLoadableValues[K, T]) Load(ctx context.Context, key K) (T, error) {
	cached, ok := v.cache.Get(key)
	if ok {
		return cached, nil
	}
	loaded, err := v.loader.Load(ctx, key)
	if err != nil {
		return loaded, err
	}
	v.cache.Add(key, loaded)
	return loaded, nil
}
