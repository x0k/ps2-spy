package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	loader_adapters "github.com/x0k/ps2-spy/internal/adapters/loader"
	sql_facility_cache "github.com/x0k/ps2-spy/internal/cache/facility/sql"
	sql_outfits_cache "github.com/x0k/ps2-spy/internal/cache/outfits/sql"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_commands "github.com/x0k/ps2-spy/internal/discord/commands"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	census_alerts_loader "github.com/x0k/ps2-spy/internal/loaders/alerts/census"
	honu_alerts_loader "github.com/x0k/ps2-spy/internal/loaders/alerts/honu"
	ps2alerts_alerts_loader "github.com/x0k/ps2-spy/internal/loaders/alerts/ps2alerts"
	voidwell_alerts_loader "github.com/x0k/ps2-spy/internal/loaders/alerts/voidwell"
	worlds_tracker_alerts_loader "github.com/x0k/ps2-spy/internal/loaders/alerts/worlds_tracker"
	census_character_ids_loader "github.com/x0k/ps2-spy/internal/loaders/character_ids/census"
	census_character_names_loader "github.com/x0k/ps2-spy/internal/loaders/character_names/census"
	census_characters_loader "github.com/x0k/ps2-spy/internal/loaders/characters/census"
	census_facility_loader "github.com/x0k/ps2-spy/internal/loaders/facility/census"
	census_outfit_ids_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_ids/census"
	census_outfit_member_ids_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids/census"
	census_outfit_tags_loader "github.com/x0k/ps2-spy/internal/loaders/outfit_tags/census"
	census_outfits_loader "github.com/x0k/ps2-spy/internal/loaders/outfits/census"
	characters_tracker_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/characters_tracker"
	fisu_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/fisu"
	honu_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/honu"
	ps2live_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/ps2live"
	saerro_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/saerro"
	sanctuary_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/sanctuary"
	voidwell_population_loader "github.com/x0k/ps2-spy/internal/loaders/population/voidwell"
	characters_tracker_trackable_online_entities_loader "github.com/x0k/ps2-spy/internal/loaders/trackable_online_entities/characters_tracker"
	census_world_map_loader "github.com/x0k/ps2-spy/internal/loaders/world_map/census"
	characters_tracker_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/characters_tracker"
	honu_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/honu"
	saerro_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/saerro"
	voidwell_world_population_loader "github.com/x0k/ps2-spy/internal/loaders/world_population/voidwell"
	worlds_tracker_world_territory_control_loader "github.com/x0k/ps2-spy/internal/loaders/world_territory_control/worlds_tracker"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
	events_module "github.com/x0k/ps2-spy/internal/modules/events"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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

	honuClient := honu.NewClient("https://wt.honu.pw", httpClient)
	fisuClient := fisu.NewClient("https://ps2.fisu.pw", httpClient)
	voidWellClient := voidwell.NewClient("https://api.voidwell.com", httpClient)
	populationClient := population.NewClient("https://agg.ps2.live", httpClient)
	saerroClient := saerro.NewClient("https://saerro.ps2.live", httpClient)
	ps2alertsClient := ps2alerts.NewClient("https://api.ps2alerts.com", httpClient)
	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.CensusServiceId, httpClient)
	sanctuaryClient := census2.NewClient("https://census.lithafalcon.cc", cfg.CensusServiceId, httpClient)

	facilityCache := sql_facility_cache.New(
		log.With(sl.Component("facility_cache")),
		storage,
	)

	characterTrackerSubsMangers := make(map[ps2_platforms.Platform]pubsub.SubscriptionsManager[characters_tracker.EventType], len(ps2_platforms.Platforms))
	charactersTrackers := make(map[ps2_platforms.Platform]*characters_tracker.CharactersTracker, len(ps2_platforms.Platforms))
	worldTrackerSubsMangers := make(map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType], len(ps2_platforms.Platforms))
	worldTrackers := make(map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker, len(ps2_platforms.Platforms))
	trackingManagers := make(map[ps2_platforms.Platform]*tracking_manager.TrackingManager, len(ps2_platforms.Platforms))
	outfitMembersSynchronizers := make(map[ps2_platforms.Platform]*outfit_members_synchronizer.OutfitMembersSynchronizer, len(ps2_platforms.Platforms))
	charactersLoaders := make(map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms))
	characterLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms))
	outfitsLoaders := make(map[ps2_platforms.Platform]loader.Multi[ps2.OutfitId, ps2.Outfit], len(ps2_platforms.Platforms))
	outfitLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit], len(ps2_platforms.Platforms))
	characterNamesLoaders := make(map[ps2_platforms.Platform]loader.Queried[[]ps2.CharacterId, []string], len(ps2_platforms.Platforms))
	characterIdsLoaders := make(map[ps2_platforms.Platform]loader.Queried[[]string, []ps2.CharacterId], len(ps2_platforms.Platforms))
	outfitTagsLoaders := make(map[ps2_platforms.Platform]loader.Queried[[]ps2.OutfitId, []string], len(ps2_platforms.Platforms))
	outfitIdsLoaders := make(map[ps2_platforms.Platform]loader.Queried[[]string, []ps2.OutfitId], len(ps2_platforms.Platforms))
	facilityLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility], len(ps2_platforms.Platforms))

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
			mt,
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

		charactersCache := expirable.NewLRU[ps2.CharacterId, ps2.Character](0, nil, 24*time.Hour)
		cachedCharactersLoader := loader.WithMultiCache(
			pl.Logger.With(sl.Component("characters_loader_cache")),
			charactersLoader,
			memory.NewMultiExpirableCache(charactersCache),
		)
		charactersLoaders[platform] = cachedCharactersLoader

		batchedCharactersLoader := loader.WithBatching(cachedCharactersLoader, 10*time.Second)
		m.Append(loader_adapters.NewBatchingService("batched_characters_loader", batchedCharactersLoader))

		cachedBatchedCharactersLoader := loader.Keyed[ps2.CharacterId, ps2.Character](
			loader.WithQueriedCache(
				pl.Logger.With(sl.Component("cached_batched_characters_loader")),
				metrics.InstrumentQueriedLoaderWithCounterMetric(
					metrics.PlatformLoadsCounterMetric(mt, metrics.CharacterPlatformLoaderName, platform),
					batchedCharactersLoader.Load,
				),
				memory.NewKeyedExpirableCache(charactersCache),
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
		worldTrackerSubsMangers[platform] = worldsTrackerPubSub
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

		trackingManager := tracking_manager.New(
			fmt.Sprintf("%s.tracking_manager", platform),
			pl.With(sl.Component("tracking_manager")),
			cachedBatchedCharactersLoader,
			func(ctx context.Context, c ps2.Character) ([]discord.Channel, error) {
				return storage.TrackingChannelsForCharacter(ctx, platform, c.Id, c.OutfitId)
			},
			func(ctx context.Context) ([]ps2.CharacterId, error) {
				return storage.AllTrackableCharacterIdsWithDuplicationsForPlatform(ctx, platform)
			},
			func(ctx context.Context, oi ps2.OutfitId) ([]ps2.CharacterId, error) {
				return storage.OutfitMembers(ctx, platform, oi)
			},
			func(ctx context.Context, oi ps2.OutfitId) ([]discord.Channel, error) {
				return storage.TrackingChannelsForOutfit(ctx, platform, oi)
			},
			func(ctx context.Context) ([]ps2.OutfitId, error) {
				return storage.AllTrackableOutfitIdsWithDuplicationsForPlatform(ctx, platform)
			},
		)
		m.Append(trackingManager)
		trackingManagers[platform] = trackingManager

		censusOutfitMemberIdsLoader := census_outfit_member_ids_loader.New(
			censusClient,
			ns,
		)

		outfitMembersSynchronizer := outfit_members_synchronizer.New(
			fmt.Sprintf("%s.outfit_members_synchronizer", platform),
			pl.With(sl.Component("outfit_members_synchronizer")),
			func(ctx context.Context) ([]ps2.OutfitId, error) {
				return storage.AllUniqueTrackableOutfitIdsForPlatform(ctx, platform)
			},
			func(ctx context.Context, oi ps2.OutfitId) (time.Time, error) {
				return storage.OutfitSynchronizedAt(ctx, platform, oi)
			},
			censusOutfitMemberIdsLoader.Load,
			func(ctx context.Context, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
				return storage.SaveOutfitMembers(ctx, platform, outfitId, members)
			},
			24*time.Hour,
		)
		m.Append(outfitMembersSynchronizer)
		outfitMembersSynchronizers[platform] = outfitMembersSynchronizer

		outfitsLoader := loader.WithMultiCache(
			log.Logger.With(sl.Component("outfits_loader_cache")),
			census_outfits_loader.New(censusClient, platform).Load,
			sql_outfits_cache.New(
				log.With(sl.Component("outfits_cache")),
				storage,
				platform,
			),
		)
		outfitsLoaders[platform] = outfitsLoader
		// We don't need batching here right now
		outfitLoaders[platform] = func(ctx context.Context, oi ps2.OutfitId) (ps2.Outfit, error) {
			outfit, err := outfitsLoader(ctx, []ps2.OutfitId{oi})
			if err != nil {
				return ps2.Outfit{}, err
			}
			return outfit[oi], nil
		}

		// TODO: Use characters cache
		characterNamesLoaders[platform] = census_character_names_loader.New(censusClient, ns).Load
		characterIdsLoaders[platform] = census_character_ids_loader.New(censusClient, ns).Load
		// TODO: Use outfits cache
		outfitTagsLoaders[platform] = census_outfit_tags_loader.New(censusClient, ns).Load
		outfitIdsLoaders[platform] = census_outfit_ids_loader.New(censusClient, ns).Load

		facilityLoaders[platform] = loader.WithKeyedCache(
			log.Logger.With(sl.Component("facilities_loader_cache")),
			census_facility_loader.New(censusClient, ns).Load,
			facilityCache,
		)
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
		"honu":      honu_population_loader.New(honuClient),
		"ps2live":   ps2live_population_loader.New(populationClient),
		"saerro":    saerro_population_loader.New(saerroClient),
		"fisu":      fisu_population_loader.New(fisuClient),
		"sanctuary": sanctuary_population_loader.New(sanctuaryClient),
		"voidwell":  voidwell_population_loader.New(voidWellClient),
	}

	worldPopulationLoaders := map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]{
		"spy": characters_tracker_world_population_loader.New(
			cfg.AppName,
			charactersTrackers,
		),
		"honu":     honu_world_population_loader.New(honuClient),
		"saerro":   saerro_world_population_loader.New(saerroClient),
		"voidwell": voidwell_world_population_loader.New(voidWellClient),
	}

	alertsLoaders := map[string]loader.Simple[meta.Loaded[ps2.Alerts]]{
		"spy": worlds_tracker_alerts_loader.New(
			log.With(sl.Component("worlds_tracker_alerts_loader")),
			cfg.AppName,
			worldTrackers,
		),
		"ps2alerts": ps2alerts_alerts_loader.New(ps2alertsClient),
		"honu":      honu_alerts_loader.New(honuClient),
		"census": census_alerts_loader.New(
			log.With(sl.Component("census_alerts_loader")),
			censusClient,
		),
		"voidwell": voidwell_alerts_loader.New(voidWellClient),
	}

	trackableOnlineEntitiesLoader := characters_tracker_trackable_online_entities_loader.New(
		storage.TrackingSettings,
		charactersTrackers,
	)

	discordMessages := discord_messages.New()
	discordCommands := discord_commands.New(
		"discord_commands",
		log.With(sl.Component("commands")),
		discordMessages,
		populationLoaders,
		[]string{"spy", "honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
		worldPopulationLoaders,
		[]string{"spy", "honu", "saerro", "voidwell"},
		worlds_tracker_world_territory_control_loader.New(
			cfg.AppName,
			worldTrackers,
		),
		alertsLoaders,
		[]string{"spy", "ps2alerts", "honu", "census", "voidwell"},
		trackableOnlineEntitiesLoader,
		func(ctx context.Context, pq discord.PlatformQuery[[]ps2.OutfitId]) (map[ps2.OutfitId]ps2.Outfit, error) {
			return outfitsLoaders[pq.Platform](ctx, pq.Value)
		},
		storage.TrackingSettings,
		func(ctx context.Context, pq discord.PlatformQuery[[]ps2.CharacterId]) ([]string, error) {
			return characterNamesLoaders[pq.Platform](ctx, pq.Value)
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]string]) ([]ps2.CharacterId, error) {
			return characterIdsLoaders[pq.Platform](ctx, pq.Value)
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]ps2.OutfitId]) ([]string, error) {
			return outfitTagsLoaders[pq.Platform](ctx, pq.Value)
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]string]) ([]ps2.OutfitId, error) {
			return outfitIdsLoaders[pq.Platform](ctx, pq.Value)
		},
		storage.SaveTrackingSettings,
		storage.SaveChannelLanguage,
	)
	m.Append(discordCommands)
	discordModule, err := discord_module.New(
		log.With(sl.Module("discord")),
		cfg.Discord.Token,
		discordCommands.Commands(),
		cfg.Discord.CommandHandlerTimeout,
		cfg.Discord.EventHandlerTimeout,
		cfg.Discord.RemoveCommands,
		characterTrackerSubsMangers,
		trackingManagers,
		discord_events.NewHandlers(
			log.With(sl.Component("discord_event_handlers")),
			discordMessages,
			characterLoaders,
			outfitLoaders,
			charactersLoaders,
			facilityLoaders,
		),
		storagePubSub,
		worldTrackerSubsMangers,
	)
	if err != nil {
		return nil, err
	}
	m.Append(discordModule)

	return m, nil
}
