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

	instrumentedEventsPublisher := metrics.InstrumentPlatformPublisher(
		mt,
		metrics.Ps2EventsPlatformPublisher,
		platform,
		eventsPublisher,
	)

	reLoginOmitter := relogin_omitter.NewReLoginOmitter(
		log.With(sl.Component("relogin_omitter")),
		instrumentedEventsPublisher,
		mt,
		platform,
	)
	m.AppendR(fmt.Sprintf("%s.relogin_omitter", platform), reLoginOmitter.Start)

	rawEventsPublisher := events.NewPublisher(reLoginOmitter)

	serviceMessagePayloadPublisher := metrics.InstrumentPlatformPublisher(
		mt,
		metrics.Ps2MessagesPlatformPublisher,
		platform,
		newServiceMessagePayloadPublisher(rawEventsPublisher),
	)

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
