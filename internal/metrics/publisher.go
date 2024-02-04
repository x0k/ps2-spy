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
	p.counter.With(prometheus.Labels{"event_type": event.Type()}).Inc()
	return p.Publish(event)
}

func instrumentPublisherCounter[E publisher.Event](counter *prometheus.CounterVec, publisher publisher.Publisher[E]) publisher.Publisher[E] {
	return &instrumentedPublisher[E]{
		Publisher: publisher,
		counter:   counter,
	}
}
