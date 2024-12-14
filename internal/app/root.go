package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	sql_facility_cache "github.com/x0k/ps2-spy/internal/cache/facility/sql"
	sql_outfits_cache "github.com/x0k/ps2-spy/internal/cache/outfits/sql"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	census_data_provider "github.com/x0k/ps2-spy/internal/data_providers/census"
	fisu_data_provider "github.com/x0k/ps2-spy/internal/data_providers/fisu"
	honu_data_provider "github.com/x0k/ps2-spy/internal/data_providers/honu"
	ps2alerts_data_provider "github.com/x0k/ps2-spy/internal/data_providers/ps2alerts"
	ps2live_data_provider "github.com/x0k/ps2-spy/internal/data_providers/ps2live"
	saerro_data_provider "github.com/x0k/ps2-spy/internal/data_providers/saerro"
	sanctuary_data_provider "github.com/x0k/ps2-spy/internal/data_providers/sanctuary"
	voidwell_data_provider "github.com/x0k/ps2-spy/internal/data_providers/voidwell"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/migrator"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
	events_module "github.com/x0k/ps2-spy/internal/modules/events"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"

	// migration tools
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	mig := migrator.New(
		log.Logger.With(sl.Component("migrator")),
		cfg.Storage.Path,
		cfg.Storage.MigrationsPath,
	)
	m.PreStartR("migrator", mig.Migrate)

	storagePubSub := pubsub.New[storage.EventType]()

	storage := sql_storage.New(
		log.With(sl.Component("storage")),
		cfg.Storage.Path,
		storagePubSub,
	)
	m.PreStartR("storage", storage.Open)
	m.PreStopR("storage", storage.Close)

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
		Transport: metrics.InstrumentTransport(
			mt,
			metrics.DefaultTransportName,
			http.DefaultTransport,
		),
	}

	censusDataProvider := census_data_provider.New(
		log.With(sl.Component("census_data_provider")),
		census2.NewClient("https://census.daybreakgames.com", cfg.CensusServiceId, httpClient),
	)
	honuDataProvider := honu_data_provider.New(
		honu.NewClient("https://wt.honu.pw", httpClient),
	)
	ps2alertsDataProvider := ps2alerts_data_provider.New(
		ps2alerts.NewClient("https://api.ps2alerts.com", httpClient),
	)
	voidwellDataProvider := voidwell_data_provider.New(
		voidwell.NewClient("https://api.voidwell.com", httpClient),
	)
	fisuDataProvider := fisu_data_provider.New(
		fisu.NewClient("https://ps2.fisu.pw", httpClient),
	)
	ps2LiveDataProvider := ps2live_data_provider.New(
		population.NewClient("https://agg.ps2.live", httpClient),
	)
	sanctuaryDataProvider := sanctuary_data_provider.New(
		census2.NewClient("https://census.lithafalcon.cc", cfg.CensusServiceId, httpClient),
	)
	saerroDataProvider := saerro_data_provider.New(
		saerro.NewClient("https://saerro.ps2.live", httpClient),
	)

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
	facilityLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility], len(ps2_platforms.Platforms))

	statsTrackerPubSub := pubsub.New[stats_tracker.EventType]()

	statsTracker := stats_tracker.New(
		log.With(sl.Component("stats_tracker")),
		statsTrackerPubSub,
		func(ctx context.Context, pq discord.PlatformQuery[ps2.CharacterId]) ([]discord.ChannelId, error) {
			manager := trackingManagers[pq.Platform]
			channels, err := manager.ChannelIdsForCharacter(ctx, pq.Value)
			if err != nil {
				return nil, err
			}
			channelIds := make([]discord.ChannelId, 0, len(channels))
			for _, channel := range channels {
				channelIds = append(channelIds, channel.ChannelId)
			}
			return channelIds, nil
		},
		storage.ChannelTrackablePlatforms,
		charactersLoaders,
		cfg.MaxTrackingDuration,
	)

	for _, platform := range ps2_platforms.Platforms {
		pl := log.With(slog.String("platform", string(platform)))
		ns := ps2_platforms.PlatformNamespace(platform)

		eventsPubSub := pubsub.New[events.EventType]()

		eventsModule, err := events_module.New(
			pl.With(sl.Module("events")),
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
			func(ctx context.Context, k []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
				return censusDataProvider.Characters(ctx, platform, k)
			},
		)

		charactersCache := expirable.NewLRU[ps2.CharacterId, ps2.Character](0, nil, 24*time.Hour)
		cachedCharactersLoader := loader.WithMultiCache(
			pl.Logger.With(sl.Component("characters_loader_cache")),
			charactersLoader,
			memory.NewMultiExpirableCache(charactersCache),
		)
		charactersLoaders[platform] = cachedCharactersLoader

		batchedCharactersLoader := loader.WithBatching(
			cachedCharactersLoader,
			10*time.Second,
			shared.ErrNotFound,
		)
		m.Append(module.NewRun(
			fmt.Sprintf("%s.batched_characters_loader", platform),
			func(ctx context.Context) error {
				batchedCharactersLoader.Start(ctx)
				return nil
			},
		))

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
			pl.With(sl.Component("characters_tracker")),
			platform,
			ps2.PlatformWorldIds[platform],
			cachedBatchedCharactersLoader,
			charactersTrackerPublisher,
			mt,
		)
		m.AppendR(
			fmt.Sprintf("%s.characters_tracker", platform),
			func(ctx context.Context) error {
				charactersTracker.Start(ctx)
				return nil
			},
		)
		charactersTrackers[platform] = charactersTracker

		worldsTrackerPubSub := pubsub.New[worlds_tracker.EventType]()

		worldsTackerPublisher := metrics.InstrumentPlatformPublisher(
			mt,
			metrics.WorldsTrackerPlatformPublisher,
			platform,
			worldsTrackerPubSub,
		)
		worldTrackerSubsMangers[platform] = worldsTrackerPubSub

		worldsTracker := worlds_tracker.New(
			pl.With(sl.Component("worlds_tracker")),
			platform,
			5*time.Minute,
			worldsTackerPublisher,
			func(ctx context.Context, wi ps2.WorldId) (ps2.WorldMap, error) {
				return censusDataProvider.WorldMap(ctx, ns, wi)
			},
		)
		m.AppendR(fmt.Sprintf("%s.worlds_tracker", platform), worldsTracker.Start)
		worldTrackers[platform] = worldsTracker

		m.Append(newEventsSubscriptionService(
			pl.With(sl.Component("events_subscription_service")),
			platform,
			m,
			eventsPubSub,
			charactersTracker,
			worldsTracker,
			statsTracker,
		))

		trackingManager := tracking_manager.New(
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
		m.AppendR(fmt.Sprintf("%s.tracking_manager", platform), trackingManager.Start)
		trackingManagers[platform] = trackingManager

		outfitMembersSynchronizer := outfit_members_synchronizer.New(
			pl.With(sl.Component("outfit_members_synchronizer")),
			func(ctx context.Context) ([]ps2.OutfitId, error) {
				return storage.AllUniqueTrackableOutfitIdsForPlatform(ctx, platform)
			},
			func(ctx context.Context, oi ps2.OutfitId) (time.Time, error) {
				return storage.OutfitSynchronizedAt(ctx, platform, oi)
			},
			func(ctx context.Context, oi ps2.OutfitId) ([]ps2.CharacterId, error) {
				return censusDataProvider.OutfitMemberIds(ctx, ns, oi)
			},
			func(ctx context.Context, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
				return storage.SaveOutfitMembers(ctx, platform, outfitId, members)
			},
			24*time.Hour,
		)
		m.AppendR(
			fmt.Sprintf("%s.outfit_members_synchronizer", platform),
			func(ctx context.Context) error {
				outfitMembersSynchronizer.Start(ctx)
				return nil
			},
		)
		outfitMembersSynchronizers[platform] = outfitMembersSynchronizer

		outfitsLoader := loader.WithMultiCache(
			log.Logger.With(sl.Component("outfits_loader_cache")),
			func(ctx context.Context, k []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, error) {
				return censusDataProvider.Outfits(ctx, platform, k)
			},
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

		facilityLoaders[platform] = loader.WithKeyedCache(
			log.Logger.With(sl.Component("facilities_loader_cache")),
			func(ctx context.Context, id ps2.FacilityId) (ps2.Facility, error) {
				return censusDataProvider.Facility(ctx, ns, id)
			},
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
		"spy": func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
			total := 0
			worlds := make([]ps2.WorldPopulation, 0)
			for _, platform := range ps2_platforms.Platforms {
				tracker, ok := charactersTrackers[platform]
				if !ok {
					log.Warn(ctx, "no population tracker for platform", slog.String("platform", string(platform)))
					continue
				}
				population := tracker.WorldsPopulation()
				total += population.Total
				worlds = append(worlds, population.Worlds...)
			}
			return meta.LoadedNow(cfg.AppName, ps2.WorldsPopulation{
				Total:  total,
				Worlds: worlds,
			}), nil
		},
		"honu":      honuDataProvider.Population,
		"ps2live":   ps2LiveDataProvider.Population,
		"saerro":    saerroDataProvider.Population,
		"fisu":      fisuDataProvider.Population,
		"sanctuary": sanctuaryDataProvider.Population,
		"voidwell":  voidwellDataProvider.Population,
	}

	worldPopulationLoaders := map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]{
		"spy": func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
			platform, ok := ps2.WorldPlatforms[worldId]
			if !ok {
				return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("unknown world %q", worldId)
			}
			population, err := charactersTrackers[platform].DetailedWorldPopulation(worldId)
			if err != nil {
				return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("getting population: %w", err)
			}
			return meta.LoadedNow(cfg.AppName, population), nil
		},
		"honu":     honuDataProvider.WorldPopulation,
		"saerro":   saerroDataProvider.WorldPopulation,
		"voidwell": voidwellDataProvider.WorldPopulation,
	}

	alertsLoaders := map[string]loader.Simple[meta.Loaded[ps2.Alerts]]{
		"spy": func(ctx context.Context) (meta.Loaded[ps2.Alerts], error) {
			alerts := make(ps2.Alerts, 0)
			for _, platform := range ps2_platforms.Platforms {
				tracker, ok := worldTrackers[platform]
				if !ok {
					log.Warn(ctx, "no alerts tracker for platform", slog.String("platform", string(platform)))
					continue
				}
				alerts = append(alerts, tracker.Alerts()...)
			}
			return meta.LoadedNow(cfg.AppName, alerts), nil
		},
		"ps2alerts": ps2alertsDataProvider.Alerts,
		"honu":      honuDataProvider.Alerts,
		"census":    censusDataProvider.Alerts,
		"voidwell":  voidwellDataProvider.Alerts,
	}

	discordModule, err := discord_module.New(
		log.With(sl.Module("discord")),
		cfg.Discord.Token,
		cfg.Discord.CommandHandlerTimeout,
		cfg.Discord.EventHandlerTimeout,
		cfg.Discord.RemoveCommands,
		characterTrackerSubsMangers,
		trackingManagers,
		storagePubSub,
		worldTrackerSubsMangers,
		populationLoaders,
		worldPopulationLoaders,
		func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.WorldTerritoryControl], error) {
			platform, ok := ps2.WorldPlatforms[worldId]
			if !ok {
				return meta.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("unknown world %q", worldId)
			}
			control, err := worldTrackers[platform].WorldTerritoryControl(ctx, worldId)
			if err != nil {
				return meta.Loaded[ps2.WorldTerritoryControl]{}, err
			}
			return meta.LoadedNow(cfg.AppName, control), nil
		},
		alertsLoaders,
		func(ctx context.Context, sq discord.SettingsQuery) (discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character], error) {
			settings, err := storage.TrackingSettings(ctx, sq)
			if err != nil {
				return discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{}, err
			}
			return charactersTrackers[sq.Platform].TrackableOnlineEntities(settings), nil
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]ps2.OutfitId]) (map[ps2.OutfitId]ps2.Outfit, error) {
			return outfitsLoaders[pq.Platform](ctx, pq.Value)
		},
		storage.TrackingSettings,
		func(ctx context.Context, pq discord.PlatformQuery[[]ps2.CharacterId]) ([]string, error) {
			return censusDataProvider.CharacterNames(ctx, ps2_platforms.PlatformNamespace(pq.Platform), pq.Value)
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]string]) ([]ps2.CharacterId, error) {
			return censusDataProvider.CharacterIds(ctx, ps2_platforms.PlatformNamespace(pq.Platform), pq.Value)
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]ps2.OutfitId]) ([]string, error) {
			return censusDataProvider.OutfitTags(ctx, ps2_platforms.PlatformNamespace(pq.Platform), pq.Value)
		},
		func(ctx context.Context, pq discord.PlatformQuery[[]string]) ([]ps2.OutfitId, error) {
			return censusDataProvider.OutfitIds(ctx, ps2_platforms.PlatformNamespace(pq.Platform), pq.Value)
		},
		storage.SaveTrackingSettings,
		storage.SaveChannelLanguage,
		characterLoaders,
		outfitLoaders,
		charactersLoaders,
		facilityLoaders,
		func(ctx context.Context, channelId discord.ChannelId) (int, error) {
			count := 0
			errs := make([]error, 0, len(ps2_platforms.Platforms))
			for _, platform := range ps2_platforms.Platforms {
				settings, err := storage.TrackingSettings(ctx, discord.SettingsQuery{
					ChannelId: channelId,
					Platform:  platform,
				})
				if err != nil {
					errs = append(errs, err)
					continue
				}
				entities := charactersTrackers[platform].TrackableOnlineEntities(settings)
				for _, outfit := range entities.Outfits {
					count += len(outfit)
				}
				count += len(entities.Characters)
			}
			return count, errors.Join(errs...)
		},
		statsTracker,
		statsTrackerPubSub,
		storage.ChannelLanguage,
	)
	if err != nil {
		return nil, err
	}
	m.Append(discordModule)

	return m, nil
}
