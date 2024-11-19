package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	loader_adapters "github.com/x0k/ps2-spy/internal/adapters/loader"
	sql_facility_cache "github.com/x0k/ps2-spy/internal/cache/facility/sql"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	discord_commands "github.com/x0k/ps2-spy/internal/discord/commands"
	discord_handlers "github.com/x0k/ps2-spy/internal/discord/handlers"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
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
	census_outfit_member_ids_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids/census"
	sql_outfit_member_ids_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids/sql"
	sql_outfit_sync_at_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_sync_at/sql"
	sql_outfit_tracking_channels_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_tracking_channels/sql"
	characters_tracker_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/characters_tracker"
	sql_trackable_character_ids_loader "github.com/x0k/ps2-spy/internal/loaders/trackable_character_ids/sql"
	sql_trackable_outfits_loader "github.com/x0k/ps2-spy/internal/loaders/trackable_outfits/sql"
	sql_trackable_outfits_with_duplication_loader "github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_with_duplication/sql"
	census_world_map_loader "github.com/x0k/ps2-spy/internal/loaders/world_map/census"
	characters_tracker_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/characters_tracker"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
	events_module "github.com/x0k/ps2-spy/internal/modules/events"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_outfit_members_saver "github.com/x0k/ps2-spy/internal/savers/outfit_members/sql"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func NewRoot(cfg *Config, log *logger.Logger) (*module.Root, error) {
	m := module.NewRoot(log.Logger)

	if cfg.Profiler.Enabled {
		m.Append(newProfilerService(cfg.Profiler.Address, m))
	}

	var mt *metrics.Metrics
	if cfg.Metrics.Enabled {
		mt = metrics.New("ps2spy")
		m.Append(newMetricsService(mt, cfg.Metrics.Address, m))
	}

	storagePubSub := pubsub.New[storage.EventType]()

	storage := sql_storage.New(
		"storage",
		log.With(sl.Component("storage")),
		cfg.Storage.Path,
		storagePubSub,
	)
	m.PreStart(module.NewHook(storage.Name(), storage.Open))
	m.PreStop(module.NewHook(storage.Name(), storage.Close))

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
		Transport: metrics.InstrumentTransport(
			mt,
			metrics.DefaultTransportName,
			http.DefaultTransport,
		),
	}

	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.CensusServiceId, httpClient)

	characterTrackingChannelsLoader := sql_character_tracking_channels_loader.New(
		storage,
	)

	characterTrackerSubsMangers := make(map[ps2_platforms.Platform]pubsub.SubscriptionsManager[characters_tracker.EventType], len(ps2_platforms.Platforms))
	charactersTrackers := make(map[ps2_platforms.Platform]*characters_tracker.CharactersTracker, len(ps2_platforms.Platforms))
	worldTrackers := make(map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker, len(ps2_platforms.Platforms))
	trackingManagers := make(map[ps2_platforms.Platform]*tracking_manager.TrackingManager, len(ps2_platforms.Platforms))
	outfitMembersSaverPublishers := make(map[ps2_platforms.Platform]pubsub.Publisher[sql_outfit_members_saver.Event], len(ps2_platforms.Platforms))
	outfitMembersSynchronizers := make(map[ps2_platforms.Platform]*outfit_members_synchronizer.OutfitMembersSynchronizer, len(ps2_platforms.Platforms))
	characterLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms))

	for _, platform := range ps2_platforms.Platforms {
		pl := log.With(slog.String("platform", string(platform)))
		ns := ps2_platforms.PlatformNamespace(platform)

		eventsPubSub := pubsub.New[events.EventType]()

		eventsModule, err := events_module.New(
			pl.With(sl.Module(fmt.Sprintf("ps2.%s.events", platform))),
			platform,
			cfg.StreamingEndpoint,
			cfg.CensusServiceId,
			eventsPubSub,
		)
		if err != nil {
			return nil, err
		}
		m.Append(eventsModule)

		charactersLoader := metrics.InstrumentMultiKeyedLoaderWithSubjectsCounter(
			metrics.PlatformLoaderSubjectsCounterMetric(mt, metrics.CharactersPlatformLoaderName, platform),
			census_characters_loader.New(
				pl.With(sl.Component("characters_loader"), slog.String("platform", string(platform))),
				censusClient,
				platform,
			).Load,
		)

		batchedCharactersLoader := loader.WithBatching(charactersLoader, 10*time.Second)
		m.Append(loader_adapters.NewBatchingService("batched_characters_loader", batchedCharactersLoader))

		cachedBatchedCharactersLoader := loader.Keyed[ps2.CharacterId, ps2.Character](
			loader.WithQueriedCache(
				pl.With(sl.Component("cached_batched_characters_loader")),
				metrics.InstrumentQueriedLoaderWithCounterMetric(
					metrics.PlatformLoadsCounterMetric(mt, metrics.CharacterPlatformLoaderName, platform),
					batchedCharactersLoader.Load,
				),
				memory.NewKeyedExpirableCache[ps2.CharacterId, ps2.Character](0, 24*time.Hour),
			),
		)
		characterLoaders[platform] = cachedBatchedCharactersLoader

		ps := pubsub.New[characters_tracker.EventType]()
		charactersTrackerPublisher := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.CharactersTrackerPlatformPublisher,
			platform,
			ps,
		)
		characterTrackerSubsMangers[platform] = ps

		charactersTracker := characters_tracker.New(
			fmt.Sprintf("%s.characters_tracker", platform),
			pl.With(sl.Component("characters_tracker")),
			platform,
			ps2.PlatformWorldIds[platform],
			cachedBatchedCharactersLoader,
			charactersTrackerPublisher,
			mt,
		)
		m.Append(charactersTracker)
		charactersTrackers[platform] = charactersTracker

		worldsTrackerPubSub := pubsub.New[worlds_tracker.EventType]()

		worldsTackerPublisher := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.WorldsTrackerPlatformPublisher,
			platform,
			worldsTrackerPubSub,
		)
		worldMapLoader := census_world_map_loader.New(
			censusClient,
			platform,
		)
		worldsTracker := worlds_tracker.New(
			fmt.Sprintf("%s.worlds_tracker", platform),
			pl.With(sl.Component("worlds_tracker")),
			platform,
			5*time.Minute,
			worldsTackerPublisher,
			worldMapLoader.Load,
		)
		m.Append(worldsTracker)
		worldTrackers[platform] = worldsTracker

		m.Append(newEventsSubscriptionService(
			pl.With(sl.Component("events_subscription_service")),
			platform,
			m,
			eventsPubSub,
			charactersTracker,
			worldsTracker,
		))

		trackableCharacterIdsLoader := sql_trackable_character_ids_loader.New(
			storage,
			platform,
		)
		sqlOutfitMemberIdsLoader := sql_outfit_member_ids_loader.New(
			storage,
			platform,
		)
		outfitTrackingChannelsLoader := sql_outfit_tracking_channels_loader.New(
			storage,
			platform,
		)
		trackableOutfitsWithDuplicationsLoader := sql_trackable_outfits_with_duplication_loader.New(
			storage,
			platform,
		)
		trackingManager := tracking_manager.New(
			fmt.Sprintf("%s.tracking_manager", platform),
			pl.With(sl.Component("tracking_manager")),
			cachedBatchedCharactersLoader,
			characterTrackingChannelsLoader,
			trackableCharacterIdsLoader,
			sqlOutfitMemberIdsLoader,
			outfitTrackingChannelsLoader,
			trackableOutfitsWithDuplicationsLoader,
		)
		m.Append(trackingManager)
		trackingManagers[platform] = trackingManager

		outfitMembersSaverPubSub := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.OutfitsMembersSaverPlatformPublisher,
			platform,
			pubsub.New[sql_outfit_members_saver.EventType](),
		)
		outfitMembersSaverPublishers[platform] = outfitMembersSaverPubSub

		trackableOutfitsLoader := sql_trackable_outfits_loader.New(
			storage,
			platform,
		)
		outfitSyncAtLoader := sql_outfit_sync_at_loader.New(
			storage,
			platform,
		)
		censusOutfitMemberIdsLoader := census_outfit_member_ids_loader.New(
			censusClient,
			ns,
		)
		outfitMembersSaver := sql_outfit_members_saver.New(
			pl.With(sl.Component("outfit_members_saver")),
			storage,
			outfitMembersSaverPubSub,
			platform,
		)
		outfitMembersSynchronizer := outfit_members_synchronizer.New(
			fmt.Sprintf("%s.outfit_members_synchronizer", platform),
			pl.With(sl.Component("outfit_members_synchronizer")),
			trackableOutfitsLoader,
			outfitSyncAtLoader,
			censusOutfitMemberIdsLoader.Load,
			outfitMembersSaver,
			24*time.Hour,
		)
		m.Append(outfitMembersSynchronizer)
		outfitMembersSynchronizers[platform] = outfitMembersSynchronizer
	}

	m.Append(newStorageEventsSubscriptionService(
		log.With(sl.Component("storage_events_subscription_service")),
		m,
		trackingManagers,
		outfitMembersSynchronizers,
		storagePubSub,
	))

	populationLoaders := map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]]{
		"spy": characters_tracker_population_loader.New(
			log.With(sl.Component("characters_tracker_population_loader")),
			cfg.AppName,
			charactersTrackers,
		),
	}
	_ = populationLoaders

	worldPopulationLoaders := map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]{
		"spy": characters_tracker_world_population_loader.New(
			cfg.AppName,
			charactersTrackers,
		),
	}
	_ = worldPopulationLoaders

	alertsLoaders := map[string]loader.Simple[meta.Loaded[ps2.Alerts]]{
		"spy": worlds_tracker_alerts_loader.New(
			log.With(sl.Component("worlds_tracker_alerts_loader")),
			cfg.AppName,
			worldTrackers,
		),
	}
	_ = alertsLoaders

	facilityCache := sql_facility_cache.New(
		log.With(sl.Component("facility_cache")),
		storage,
	)
	_ = facilityCache

	discordMessages := discord_messages.New()
	discordModule, err := discord_module.New(
		log.With(sl.Module("discord")),
		cfg.Discord.Token,
		discord_commands.New(discordMessages),
		cfg.Discord.CommandHandlerTimeout,
		cfg.Discord.EventHandlerTimeout,
		cfg.Discord.RemoveCommands,
		characterTrackerSubsMangers,
		trackingManagers,
		discord_handlers.New(
			discordMessages,
			characterLoaders,
		),
	)
	if err != nil {
		return nil, err
	}
	m.Append(discordModule)

	return m, nil
}
