package metrics

import (
	"errors"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
)

var ErrInvalidEventType = errors.New("invalid event type")

type instrumentedPublisher[T pubsub.EventType] struct {
	pubsub.Publisher[pubsub.Event[T]]
	counter *prometheus.CounterVec
}

func (p *instrumentedPublisher[T]) Publish(event pubsub.Event[T]) error {
	err := p.Publisher.Publish(event)
	var eventType string
	switch v := any(event.Type()).(type) {
	case int:
		eventType = strconv.Itoa(v)
	case string:
		eventType = v
	default:
		return ErrInvalidEventType
	}
	labels := prometheus.Labels{
		"event_type": eventType,
		"status":     string(SuccessStatus),
	}
	if err != nil {
		labels["status"] = string(ErrorStatus)
	}
	p.counter.With(labels).Inc()
	return err
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
