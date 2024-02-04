package publisher

import (
	"errors"
	"sync"
)

var ErrUnknownHandler = errors.New("unknown handler")

type Abstract[E any] interface {
	Publish(event E) error
}

type Event interface {
	Type() string
}

type Handler[E Event] interface {
	Type() string
	Handle(event E)
}

type SubscriptionsManager[E Event] interface {
	AddHandler(h Handler[E]) func()
}

type Publisher[E Event] interface {
	Abstract[E]
	SubscriptionsManager[E]
}

type publisher[E Event] struct {
	handlersMu sync.RWMutex
	handlers   map[string][]Handler[E]
}

func New[E Event]() *publisher[E] {
	return &publisher[E]{
		handlers: map[string][]Handler[E]{},
	}
}

func (p *publisher[E]) removeHandler(eventType string, h Handler[E]) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	for i, v := range p.handlers[eventType] {
		if v == h {
			p.handlers[eventType] = append(p.handlers[eventType][:i], p.handlers[eventType][i+1:]...)
			return
		}
	}
}

func (p *publisher[E]) AddHandler(h Handler[E]) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[h.Type()] = append(p.handlers[h.Type()], h)
	return func() {
		p.removeHandler(h.Type(), h)
	}
}

func (p *publisher[E]) Publish(event E) error {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	for _, h := range p.handlers[event.Type()] {
		h.Handle(event)
	}
	return nil
}
