package ps2_platform_module

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	loader_adapters "github.com/x0k/ps2-spy/internal/adapters/loader"
	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	census_characters_loader "github.com/x0k/ps2-spy/internal/loaders/characters/census"
	"github.com/x0k/ps2-spy/internal/metrics"
	ps2_events_module "github.com/x0k/ps2-spy/internal/modules/ps2/platform/events"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type Config struct {
	HttpClient        *http.Client
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
}

func New(log *logger.Logger, mt *metrics.Metrics, cfg *Config) (*module.Module, error) {
	m := module.New(log.Logger.With(slog.String("module", "ps2")), "ps2")

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

	charactersLoader := metrics.InstrumentMultiKeyedLoaderWithSubjectsCounter(
		metrics.PlatformLoaderSubjectsCounterMetric(mt, metrics.CharactersPlatformLoaderName, cfg.Platform),
		census_characters_loader.New(
			log.With(sl.Component("characters_loader"), slog.String("platform", string(cfg.Platform))),
			censusClient,
			cfg.Platform,
		).Load,
	)

	batchedCharactersLoader := loader.WithBatching(charactersLoader, 10*time.Second)
	m.Append(loader_adapters.NewBatchingService("batched_characters_loader", batchedCharactersLoader))

	cachedBatchedCharactersLoader := loader.WithQueriedCache(
		log.With(sl.Component("cached_batched_characters_loader")),
		metrics.InstrumentQueriedLoaderWithCounterMetric(
			metrics.PlatformLoadsCounterMetric(mt, metrics.CharacterPlatformLoaderName, cfg.Platform),
			batchedCharactersLoader.Load,
		),
		memory.NewKeyedExpirableCache[ps2.CharacterId, ps2.Character](0, 24*time.Hour),
	)

	charactersTrackerPubSub := pubsub.New[characters_tracker.EventType]()

	charactersTrackerPublisher := metrics.InstrumentPlatformPublisher(
		mt,
		metrics.CharactersTrackerPlatformPublisher,
		cfg.Platform,
		charactersTrackerPubSub,
	)

	charactersTracker := characters_tracker.New(
		log,
		cfg.Platform,
		[]ps2.WorldId{},
		loader.Keyed[ps2.CharacterId, ps2.Character](cachedBatchedCharactersLoader),
		charactersTrackerPublisher,
		mt,
	)
	m.Append(newCharactersTrackerService(cfg.Platform, charactersTracker))

	m.Append(newEventsSubscriptionService(cfg.Platform, m, eventsPubSub, charactersTracker))

	worldsTrackerPubSub := pubsub.New[worlds_tracker.EventType]()

	worldsTackerPublisher := metrics.InstrumentPlatformPublisher(
		mt,
		metrics.WorldsTrackerPlatformPublisher,
		cfg.Platform,
		worldsTrackerPubSub,
	)

	// populationLoader := characters_tracker_population_loader.NewCharactersTrackerLoader(
	// 	log,
	// 	cfg.BotName,
	// )

	return m, nil
}
