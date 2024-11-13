package pubsub

import (
	"errors"
)

type BufferedPublisher[T EventType] struct {
	Publisher[T]
	buffer []Event[T]
}

func NewBufferedPublisher[T EventType](pub Publisher[T], estimatedEventsCount int) *BufferedPublisher[T] {
	return &BufferedPublisher[T]{
		Publisher: pub,
		buffer:    make([]Event[T], 0, estimatedEventsCount),
	}
}

func (p *BufferedPublisher[T]) Publish(event Event[T]) error {
	p.buffer = append(p.buffer, event)
	return nil
}

func (p *BufferedPublisher[T]) Flush() error {
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
