package containers

import (
	"fmt"
	"log"
	"time"
)

var ErrAllFallbacksFailed = fmt.Errorf("all fallbacks failed")

type Fallbacks[T any] struct {
	// TODO: Replace with proper logger
	name                 string
	entities             map[string]T
	priority             []string
	lastSuccessfulEntity *ExpiableValue[string]
}

func NewFallbacks[T any](name string, entities map[string]T, priority []string, ttl time.Duration) *Fallbacks[T] {
	return &Fallbacks[T]{
		name:                 name,
		entities:             entities,
		priority:             priority,
		lastSuccessfulEntity: NewExpiableValue[string](ttl),
	}
}

func (f *Fallbacks[T]) Start() {
	go f.lastSuccessfulEntity.StartExpiration()
}

func (f *Fallbacks[T]) Stop() {
	f.lastSuccessfulEntity.StopExpiration()
}

func (f *Fallbacks[T]) Exec(executor func(T) error) error {
	if name, ok := f.lastSuccessfulEntity.Read(); ok {
		entity, ok := f.entities[name]
		if ok {
			err := executor(entity)
			if err == nil {
				return nil
			}
			log.Printf("[%s] Last successful entity %q failed: %q", f.name, name, err)
		} else {
			log.Printf("[%s] Last successful entity %q not found", f.name, name)
		}
	}
	for _, name := range f.priority {
		entity, ok := f.entities[name]
		if !ok {
			log.Printf("[%s] Fallback entity %q not found", f.name, name)
			continue
		}
		err := executor(entity)
		if err != nil {
			log.Printf("[%s] Fallback entity %q failed: %q", f.name, name, err)
			continue
		}
		f.lastSuccessfulEntity.Write(name)
		return nil
	}
	return fmt.Errorf("%s: %w", f.name, ErrAllFallbacksFailed)
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
