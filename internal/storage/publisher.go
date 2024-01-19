package storage

import (
	"errors"
	"sync"
)

var ErrUnknownEvent = errors.New("unknown event")

type Publisher struct {
	handlersMu sync.RWMutex
	handlers   map[string][]eventHandler
}

func NewPublisher() *Publisher {
	return &Publisher{
		handlers: map[string][]eventHandler{},
	}
}

func (p *Publisher) removeHandler(eventType string, h eventHandler) {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	for i, v := range p.handlers[eventType] {
		if v == h {
			p.handlers[eventType] = append(p.handlers[eventType][:i], p.handlers[eventType][i+1:]...)
			return
		}
	}
}

func (p *Publisher) addHandler(h eventHandler) func() {
	p.handlersMu.Lock()
	defer p.handlersMu.Unlock()
	p.handlers[h.Type()] = append(p.handlers[h.Type()], h)
	return func() {
		p.removeHandler(h.Type(), h)
	}
}

func (p *Publisher) AddHandler(h any) (func(), error) {
	handler := handlerForInterface(h)
	if handler == nil {
		return nil, ErrUnknownEvent
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
