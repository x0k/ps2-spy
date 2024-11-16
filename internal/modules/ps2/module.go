package ps2_module

import (
	"log/slog"
	"net/http"
	"time"

	loader_adapters "github.com/x0k/ps2-spy/internal/adapters/loader"
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
	characters_tracker_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/characters_tracker"
	census_world_map_loader "github.com/x0k/ps2-spy/internal/loaders/world_map/census"
	characters_tracker_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/characters_tracker"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	ps2_events_module "github.com/x0k/ps2-spy/internal/modules/ps2/events"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type Config struct {
	HttpClient        *http.Client
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
	AppName           string
}

func New(log *logger.Logger, mt *metrics.Metrics, cfg *Config) (*module.Module, error) {
	log = log.With(slog.String("module", "ps2"))
	m := module.New(log.Logger, "ps2")

	platformCharactersTrackers := make(map[ps2_platforms.Platform]*characters_tracker.CharactersTracker, len(ps2_platforms.Platforms))
	platformWorldTrackers := make(map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker, len(ps2_platforms.Platforms))

	for _, platform := range ps2_platforms.Platforms {
		pl := log.With(slog.String("platform", string(platform)))

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
			pl.With(sl.Component("cached_batched_characters_loader")),
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
			pl.With(sl.Component("characters_tracker")),
			platform,
			ps2.PlatformWorldIds[platform],
			loader.Keyed[ps2.CharacterId, ps2.Character](cachedBatchedCharactersLoader),
			charactersTrackerPublisher,
			mt,
		)
		m.Append(newCharactersTrackerService(cfg.Platform, charactersTracker))
		platformCharactersTrackers[platform] = charactersTracker

		worldsTrackerPubSub := pubsub.New[worlds_tracker.EventType]()

		worldsTackerPublisher := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.WorldsTrackerPlatformPublisher,
			cfg.Platform,
			worldsTrackerPubSub,
		)
		worldMapLoader := census_world_map_loader.New(
			censusClient,
			platform,
		)
		worldsTracker := worlds_tracker.New(
			pl.With(sl.Component("worlds_tracker")),
			platform,
			5*time.Minute,
			worldsTackerPublisher,
			worldMapLoader.Load,
		)
		m.Append(newWorldsTrackerService(cfg.Platform, worldsTracker))
		platformWorldTrackers[platform] = worldsTracker

		m.Append(newEventsSubscriptionService(
			log.With(sl.Component("events_subscription_service")),
			cfg.Platform,
			m,
			eventsPubSub,
			charactersTracker,
			worldsTracker,
		))
	}

	populationLoaders := map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]]{
		"spy": characters_tracker_population_loader.New(
			log.With(sl.Component("characters_tracker_population_loader")),
			cfg.AppName,
			platformCharactersTrackers,
		),
	}

	worldPopulationLoaders := map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]{
		"spy": characters_tracker_world_population_loader.New(
			cfg.AppName,
			platformCharactersTrackers,
		),
	}

	return m, nil
}
