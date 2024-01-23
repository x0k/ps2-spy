package loaders

import (
	"context"
	"fmt"
	"time"
)

var ErrNotFound = fmt.Errorf("not found")

type Loaded[T any] struct {
	Value     T
	Source    string
	UpdatedAt time.Time
}

func LoadedNow[T any](source string, value T) Loaded[T] {
	return Loaded[T]{
		Value:     value,
		Source:    source,
		UpdatedAt: time.Now(),
	}
}

type Loader[T any] interface {
	Load(ctx context.Context) (T, error)
}

type QueriedLoader[Q any, T any] interface {
	Load(ctx context.Context, query Q) (T, error)
}

type KeyedLoader[K comparable, T any] QueriedLoader[K, T]
