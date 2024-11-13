package event

import (
	"errors"

	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

type BufferedPublisher struct {
	Publisher
	buffer []Event
}

func NewBufferedPublisher(pub Publisher, estimatedEventsCount int) *BufferedPublisher {
	return &BufferedPublisher{
		Publisher: pub,
		buffer:    make([]Event, 0, estimatedEventsCount),
	}
}

func (p *BufferedPublisher) Publish(event pubsub.Event[Type]) error {
	p.buffer = append(p.buffer, event)
	return nil
}

func (p *BufferedPublisher) Flush() error {
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
