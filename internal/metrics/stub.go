package metrics

import (
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type stub struct{}

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
