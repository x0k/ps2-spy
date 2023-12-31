package containers

import (
	"context"
	"time"
)

type valueLoader[T any] interface {
	Load(ctx context.Context) (T, error)
}

type LoadableValue[T any] struct {
	value  *ExpiableValue[T]
	loader valueLoader[T]
}

func NewLoadableValue[T any](loader valueLoader[T], ttl time.Duration) *LoadableValue[T] {
	return &LoadableValue[T]{
		value:  NewExpiableValue[T](ttl),
		loader: loader,
	}
}

func (v *LoadableValue[T]) StartExpiration() {
	v.value.StartExpiration()
}

func (v *LoadableValue[T]) StopExpiration() {
	v.value.StopExpiration()
}

func (v *LoadableValue[T]) Load(ctx context.Context) (T, error) {
	return v.value.Load(func() (T, error) {
		value, err := v.loader.Load(ctx)
		if err != nil {
			return value, err
		}
		return value, nil
	})
}