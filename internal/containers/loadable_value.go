package containers

import (
	"context"
	"time"
)

type loader[T any] interface {
	Load(ctx context.Context) (T, error)
}

type LoadableValue[T any] struct {
	value     *ExpiableValue[T]
	provider  loader[T]
	updatedAt time.Time
}

func NewLoadableValue[T any](provider loader[T], ttl time.Duration) *LoadableValue[T] {
	return &LoadableValue[T]{
		value:     NewExpiableValue[T](ttl),
		provider:  provider,
		updatedAt: time.Now(),
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
		value, err := v.provider.Load(ctx)
		if err != nil {
			return value, err
		}
		v.updatedAt = time.Now()
		return value, nil
	})
}

func (v *LoadableValue[T]) UpdatedAt() time.Time {
	return v.updatedAt
}
