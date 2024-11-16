package ps2_events_module

import (
	"fmt"
	"log/slog"

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
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
}

func New(log *logger.Logger, cfg *Config, eventsPublisher pubsub.Publisher[events.Event]) (*module.Module, error) {
	m := module.New(log.Logger.With(slog.String("module", "ps2.platform.events")), fmt.Sprintf("platform.%s", cfg.Platform))

	reLoginOmitter := relogin_omitter.NewReLoginOmitter(
		log.With(sl.Component("relogin_omitter")),
		eventsPublisher,
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
