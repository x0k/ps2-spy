package ps2_module

import (
	"fmt"
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
	worlds_tracker_alerts_loader "github.com/x0k/ps2-spy/internal/loaders/alerts/worlds_tracker"
	sql_character_tracking_channels_loader "github.com/x0k/ps2-spy/internal/loaders/character_tracking_channels/sql"
	census_characters_loader "github.com/x0k/ps2-spy/internal/loaders/characters/census"
	sql_outfit_member_ids_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids/sql"
	sql_outfit_tracking_channels_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_tracking_channels/sql"
	characters_tracker_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/characters_tracker"
	sql_trackable_character_ids_loader "github.com/x0k/ps2-spy/internal/loaders/trackable_character_ids/sql"
	sql_trackable_outfits_with_duplication_loader "github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_with_duplication/sql"
	census_world_map_loader "github.com/x0k/ps2-spy/internal/loaders/world_map/census"
	characters_tracker_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/characters_tracker"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	ps2_events_module "github.com/x0k/ps2-spy/internal/modules/ps2/events"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

type Config struct {
	Log               *logger.Logger
	Metrics           *metrics.Metrics
	Storage           *sql_storage.Storage
	HttpClient        *http.Client
	Platform          ps2_platforms.Platform
	StreamingEndpoint string
	CensusServiceId   string
	AppName           string
}

func New(cfg *Config) (*module.Module, error) {
	log := cfg.Log
	mt := cfg.Metrics
	m := module.New(log.Logger, "ps2")

	characterTrackingChannelsLoader := sql_character_tracking_channels_loader.New(
		cfg.Storage,
	)

	charactersTrackers := make(map[ps2_platforms.Platform]*characters_tracker.CharactersTracker, len(ps2_platforms.Platforms))
	worldTrackers := make(map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker, len(ps2_platforms.Platforms))
	trackingManagers := make(map[ps2_platforms.Platform]*tracking_manager.TrackingManager, len(ps2_platforms.Platforms))

	for _, platform := range ps2_platforms.Platforms {
		pl := log.With(slog.String("platform", string(platform)))

		eventsPubSub := pubsub.New[events.EventType]()

		eventsModule, err := ps2_events_module.New(&ps2_events_module.Config{
			Log:               pl.With(slog.String("module", fmt.Sprintf("ps2.%s.events", platform))),
			EventsPublisher:   eventsPubSub,
			Platform:          cfg.Platform,
			StreamingEndpoint: cfg.StreamingEndpoint,
			CensusServiceId:   cfg.CensusServiceId,
		})
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
		charactersTrackers[platform] = charactersTracker

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
		worldTrackers[platform] = worldsTracker

		trackableCharacterIdsLoader := sql_trackable_character_ids_loader.New(
			cfg.Storage,
			platform,
		)
		outfitMemberIdsLoader := sql_outfit_member_ids_loader.New(
			cfg.Storage,
			platform,
		)
		outfitTrackingChannelsLoader := sql_outfit_tracking_channels_loader.New(
			cfg.Storage,
			platform,
		)
		trackableOutfitsLoader := sql_trackable_outfits_with_duplication_loader.New(
			cfg.Storage,
			platform,
		)
		trackingManager := tracking_manager.New(
			log,
			loader.Keyed[ps2.CharacterId, ps2.Character](cachedBatchedCharactersLoader),
			characterTrackingChannelsLoader,
			trackableCharacterIdsLoader,
			outfitMemberIdsLoader,
			outfitTrackingChannelsLoader,
			trackableOutfitsLoader,
		)
		m.Append(newTrackingManagerService(cfg.Platform, trackingManager))
		trackingManagers[platform] = trackingManager

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
			charactersTrackers,
		),
	}

	worldPopulationLoaders := map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]{
		"spy": characters_tracker_world_population_loader.New(
			cfg.AppName,
			charactersTrackers,
		),
	}

	alertsLoaders := map[string]loader.Simple[meta.Loaded[ps2.Alerts]]{
		"spy": worlds_tracker_alerts_loader.New(
			log.With(sl.Component("worlds_tracker_alerts_loader")),
			cfg.AppName,
			worldTrackers,
		),
	}

	return m, nil
}
