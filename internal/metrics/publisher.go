package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
)

type instrumentedPublisher[E publisher.Event] struct {
	publisher.Publisher[E]
	counter *prometheus.CounterVec
}

func (p *instrumentedPublisher[E]) Publish(event E) error {
	err := p.Publisher.Publish(event)
	labels := prometheus.Labels{
		"event_type": event.Type(),
		"status":     string(SuccessStatus),
	}
	if err != nil {
		labels["status"] = string(ErrorStatus)
	}
	p.counter.With(labels).Inc()
	return err
}

func instrumentPublisher[E publisher.Event](counter *prometheus.CounterVec, publisher publisher.Publisher[E]) publisher.Publisher[E] {
	return &instrumentedPublisher[E]{
		Publisher: publisher,
		counter:   counter,
	}
}
