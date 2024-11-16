package ps2_events_module

import (
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/relogin_omitter"
)

type Config struct {
	Log               *logger.Logger
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
	EventsPublisher   pubsub.Publisher[events.Event]
}

func New(cfg *Config) (*module.Module, error) {
	log := cfg.Log
	m := module.New(log.Logger, "ps2.events")

	reLoginOmitter := relogin_omitter.NewReLoginOmitter(
		log.With(sl.Component("relogin_omitter")),
		cfg.EventsPublisher,
	)
	m.Append(NewReLoginOmitterService(cfg.Platform, reLoginOmitter))

	rawEventsPublisher := events.NewPublisher(reLoginOmitter)

	serviceMessagePayloadPublisher := newServiceMessagePayloadPublisher(rawEventsPublisher)

	streamingPublisher := streaming.NewPublisher(serviceMessagePayloadPublisher)

	streamingClient := streaming.NewClient(
		cfg.StreamingEndpoint,
		ps2_platforms.PlatformEnvironment(cfg.Platform),
		cfg.CensusServiceId,
		streamingPublisher,
	)
	m.Append(newStreamingClientService(log, cfg.Platform, streamingClient))

	return m, nil
}
