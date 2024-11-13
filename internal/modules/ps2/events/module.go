package events_module

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/relogin_omitter"
)

type Config struct {
	Platform          platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
}

func New(log *logger.Logger, cfg *Config, eventsPublisher pubsub.Publisher[events.Event]) (*module.Module, error) {
	m := module.New(log.Logger, fmt.Sprintf("platform.%s", cfg.Platform))

	reLoginOmitter := relogin_omitter.NewReLoginOmitter(log, eventsPublisher)
	m.Append(NewReLoginOmitterService(cfg.Platform, reLoginOmitter))

	rawEventsPublisher := events.NewPublisher(reLoginOmitter)

	serviceMessagePayloadPublisher := newServiceMessagePayloadPublisher(rawEventsPublisher)

	streamingPublisher := streaming.NewPublisher(serviceMessagePayloadPublisher)

	streamingClient := streaming.NewClient(
		cfg.StreamingEndpoint,
		platforms.PlatformEnvironment(cfg.Platform),
		cfg.CensusServiceId,
		streamingPublisher,
	)
	m.Append(newStreamingClientService(log, cfg.Platform, streamingClient))

	return m, nil
}
