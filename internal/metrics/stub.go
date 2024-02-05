package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type stub struct{}

// InstrumentTransport implements Metrics.
func (*stub) InstrumentTransport(_ TransportName, t http.RoundTripper) http.RoundTripper {
	return t
}

// PlatformLoaderSubjectsCounterMetric implements Metrics.
func (*stub) PlatformLoaderSubjectsCounterMetric(PlatformLoaderName, platforms.Platform) *prometheus.GaugeVec {
	return nil
}

// PlatformLoaderInFlightMetric implements Metrics.
func (*stub) PlatformLoaderInFlightMetric(PlatformLoaderName, platforms.Platform) *prometheus.Gauge {
	return nil
}

// PlatformLoadsCounterMetric implements Metrics.
func (*stub) PlatformLoadsCounterMetric(PlatformLoaderName, platforms.Platform) *prometheus.CounterVec {
	return nil
}

// PlatformLoadsCounter implements Metrics.
func (*stub) PlatformLoadsCounter(PlatformLoaderName, platforms.Platform) *prometheus.CounterVec {
	return nil
}

// InstrumentPlatformPublisher implements Metrics.
func (*stub) InstrumentPlatformPublisher(_ PlatformPublisherName, _ platforms.Platform, p publisher.Publisher[publisher.Event]) publisher.Publisher[publisher.Event] {
	return p
}

// InstrumentPublisher implements Metrics.
func (*stub) InstrumentPublisher(_ PublisherName, p publisher.Publisher[publisher.Event]) publisher.Publisher[publisher.Event] {
	return p
}

func NewStub() Metrics {
	return &stub{}
}
