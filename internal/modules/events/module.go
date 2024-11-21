package events_module

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/relogin_omitter"
)

func New(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	streamingEndpoint string,
	censusServiceId string,
	eventsPublisher pubsub.Publisher[events.Event],
	mt *metrics.Metrics,
) (*module.Module, error) {
	m := module.New(log.Logger, "ps2.events")

	reLoginOmitter := relogin_omitter.NewReLoginOmitter(
		fmt.Sprintf("%s.relogin_omitter", platform),
		log.With(sl.Component("relogin_omitter")),
		eventsPublisher,
		mt,
		platform,
	)
	m.Append(reLoginOmitter)

	rawEventsPublisher := events.NewPublisher(reLoginOmitter)

	serviceMessagePayloadPublisher := newServiceMessagePayloadPublisher(rawEventsPublisher)

	streamingPublisher := streaming.NewPublisher(serviceMessagePayloadPublisher)

	streamingClient := streaming.NewClient(
		streamingEndpoint,
		ps2_platforms.PlatformEnvironment(platform),
		censusServiceId,
		streamingPublisher,
	)
	m.Append(newStreamingClientService(log, platform, streamingClient))

	return m, nil
}
