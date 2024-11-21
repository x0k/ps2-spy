package pubsub

import (
	"errors"
)

type BufferedPublisher[E any] struct {
	Publisher[E]
	buffer []E
}

func NewBufferedPublisher[E any](pub Publisher[E], estimatedEventsCount int) *BufferedPublisher[E] {
	return &BufferedPublisher[E]{
		Publisher: pub,
		buffer:    make([]E, 0, estimatedEventsCount),
	}
}

func (p *BufferedPublisher[E]) Publish(event E) error {
	p.buffer = append(p.buffer, event)
	return nil
}

func (p *BufferedPublisher[E]) Flush() error {
	errs := make([]error, 0, len(p.buffer))
	for _, event := range p.buffer {
		err := p.Publisher.Publish(event)
		if err != nil {
			errs = append(errs, err)
		}
	}
	p.buffer = p.buffer[:0]
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
