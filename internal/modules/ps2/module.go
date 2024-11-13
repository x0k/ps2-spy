package ps2_module

import (
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	events_module "github.com/x0k/ps2-spy/internal/modules/ps2/events"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Config struct {
	Platform          platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
}

func New(log *logger.Logger, cfg *Config) (*module.Module, error) {
	m := module.New(log.Logger, "ps2")

	eventsPubSub := pubsub.New[events.EventType]()

	eventsModule, err := events_module.New(log, &events_module.Config{}, eventsPubSub)
	if err != nil {
		return nil, err
	}
	m.Append(eventsModule)

	return m, nil
}
