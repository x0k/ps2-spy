package platform_module

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

func New(log *logger.Logger, cfg *Config) (*module.Module, error) {
	m := module.New(log.Logger, fmt.Sprintf("platform.%s", cfg.Platform))

	cleanEventsPubSub := pubsub.New[events.EventType]()

	reLoginOmitter := relogin_omitter.NewReLoginOmitter(log, cleanEventsPubSub)
	m.Append(NewReLoginOmitterService(cfg.Platform, reLoginOmitter))

	eventsPublisher := events.NewPublisher(reLoginOmitter)

	serviceMessagePayloadPublisher := newServiceMessagePayloadPublisher(eventsPublisher)

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
