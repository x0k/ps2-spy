package ps2_module

import (
	"context"
	"log/slog"
	"net/http"

	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_events_module "github.com/x0k/ps2-spy/internal/modules/ps2/events"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Config struct {
	HttpClient        *http.Client
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
}

func New(log *logger.Logger, cfg *Config) (*module.Module, error) {
	m := module.New(log.Logger, "ps2")

	eventsPubSub := pubsub.New[events.EventType]()

	eventsModule, err := ps2_events_module.New(log, &ps2_events_module.Config{
		Platform:          cfg.Platform,
		StreamingEndpoint: cfg.StreamingEndpoint,
		CensusServiceId:   cfg.CensusServiceId,
	}, eventsPubSub)
	if err != nil {
		return nil, err
	}
	m.Append(eventsModule)

	m.Append(module.NewService("events_test", func(ctx context.Context) error {
		event := pubsub_adapters.Subscribe[events.EventType, events.PlayerLogin](m, eventsPubSub)
		for {
			select {
			case <-ctx.Done():
				return nil
			case e := <-event:
				log.Debug(ctx, "player login", slog.String("playerName", e.CharacterID))
			}
		}
	}))

	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.CensusServiceId, cfg.HttpClient)
	_ = censusClient

	return m, nil
}
