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

type Handler interface {
	Type() string
	Handle(event Event)
}

type Publisher struct {
	handlersMu  sync.RWMutex
	handlers    map[string][]Handler
	castHandler func(any) Handler
}

func New(castHandler func(any) Handler) *Publisher {
	return &Publisher{
		handlers:    map[string][]Handler{},
		castHandler: castHandler,
	}
}

func (p *Publisher) removeHandler(eventType string, h Handler) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	for i, v := range p.handlers[eventType] {
		if v == h {
			p.handlers[eventType] = append(p.handlers[eventType][:i], p.handlers[eventType][i+1:]...)
			return
		}
	}
}

func (p *Publisher) addHandler(h Handler) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[h.Type()] = append(p.handlers[h.Type()], h)
	return func() {
		p.removeHandler(h.Type(), h)
	}
}

func (p *Publisher) AddHandler(h any) (func(), error) {
	handler := p.castHandler(h)
	if handler == nil {
		return nil, ErrUnknownHandler
	}
	return p.addHandler(handler), nil
}

func (p *Publisher) Publish(event Event) error {
	p.handlersMu.RLock()
	defer p.handlersMu.RUnlock()
	for _, h := range p.handlers[event.Type()] {
		h.Handle(event)
	}
	return nil
}
