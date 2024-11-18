package pubsub

import (
	"errors"
	"sync"
)

var ErrUnknownHandler = errors.New("unknown handler")

type EventType interface {
	~int | ~string
}

type Event[T EventType] interface {
	Type() T
}

type Handler[T EventType] interface {
	Type() T
	Handle(event Event[T]) error
}

type SubscriptionsManager[T EventType] interface {
	AddHandler(h Handler[T]) func()
}

type Publisher[E any] interface {
	Publish(event E) error
}

type pubSub[T EventType] struct {
	handlersMu sync.RWMutex
	handlers   map[T][]Handler[T]
}

func New[T EventType]() *pubSub[T] {
	return &pubSub[T]{
		handlers: map[T][]Handler[T]{},
	}
}

func (p *pubSub[T]) removeHandler(h Handler[T]) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	eventType := h.Type()
	for i, v := range p.handlers[eventType] {
		if v == h {
			p.handlers[eventType] = append(p.handlers[eventType][:i], p.handlers[eventType][i+1:]...)
			return
		}
	}
}

func (p *pubSub[T]) AddHandler(h Handler[T]) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[h.Type()] = append(p.handlers[h.Type()], h)
	return func() {
		p.removeHandler(h)
	}
}

func (p *pubSub[T]) Publish(event Event[T]) error {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	for _, h := range p.handlers[event.Type()] {
		if err := h.Handle(event); err != nil {
			return err
		}
	}
	return nil
}
