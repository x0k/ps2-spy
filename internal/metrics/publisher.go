package metrics

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrInvalidEventType = errors.New("invalid event type")

type instrumentedPublisher[T pubsub.EventType] struct {
	pubsub.Publisher[pubsub.Event[T]]
	counter *prometheus.CounterVec
}

func (p *instrumentedPublisher[T]) Publish(event pubsub.Event[T]) {
	p.Publisher.Publish(event)
	labels := prometheus.Labels{
		"event_type": string(event.Type()),
		"status":     string(SuccessStatus),
	}
	p.counter.With(labels).Inc()
}

func newInstrumentPublisher[T pubsub.EventType](
	counter *prometheus.CounterVec,
	publisher pubsub.Publisher[pubsub.Event[T]],
) pubsub.Publisher[pubsub.Event[T]] {
	return &instrumentedPublisher[T]{
		Publisher: publisher,
		counter:   counter,
	}
}
