package containers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

var ErrAllFallbacksFailed = fmt.Errorf("all fallbacks failed")

type Fallbacks[T any] struct {
	log                  *slog.Logger
	entities             map[string]T
	priority             []string
	lastSuccessfulEntity *Expirable[string]
}

func NewFallbacks[T any](log *slog.Logger, entities map[string]T, priority []string, ttl time.Duration) *Fallbacks[T] {
	return &Fallbacks[T]{
		log:                  log,
		entities:             entities,
		priority:             priority,
		lastSuccessfulEntity: NewExpirable[string](ttl),
	}
}

func (f *Fallbacks[T]) Start(ctx context.Context) {
	f.lastSuccessfulEntity.Start(ctx)
}

func (f *Fallbacks[T]) Exec(executor func(T) error) error {
	if name, ok := f.lastSuccessfulEntity.Get(); ok {
		entity, ok := f.entities[name]
		if ok {
			err := executor(entity)
			if err == nil {
				return nil
			}
			f.log.Debug("[ERROR] last successful fallback failed", slog.String("fallback", name), sl.Err(err))
		} else {
			f.log.Warn("last successful fallback not found", slog.String("fallback", name))
		}
	}
	for _, name := range f.priority {
		entity, ok := f.entities[name]
		if !ok {
			f.log.Warn("fallback not found", slog.String("fallback", name))
			continue
		}
		err := executor(entity)
		if err != nil {
			f.log.Debug("[ERROR] fallback failed", slog.String("fallback", name), sl.Err(err))
			continue
		}
		f.lastSuccessfulEntity.Set(name)
		return nil
	}
	return ErrAllFallbacksFailed
}

func ExecFallback[T any, R any](fallbacks *Fallbacks[T], executor func(T) (R, error)) (R, error) {
	var result R
	var err error
	err = fallbacks.Exec(func(entity T) error {
		result, err = executor(entity)
		return err
	})
	return result, err
}
