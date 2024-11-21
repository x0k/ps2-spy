package meta

import (
	"time"
)

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
