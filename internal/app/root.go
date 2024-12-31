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
	discord_commands "github.com/x0k/ps2-spy/internal/discord/commands"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
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
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_census_characters_repo "github.com/x0k/ps2-spy/internal/ps2/census_characters_repo"
	ps2_census_outfits_repo "github.com/x0k/ps2-spy/internal/ps2/census_outfits_repo"
	ps2_outfit_members_synchronizer "github.com/x0k/ps2-spy/internal/ps2/outfit_members_synchronizer"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	ps2_storage_outfits_repo "github.com/x0k/ps2-spy/internal/ps2/storage_outfits_repo"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	stats_tracker_storage_tasks_repo "github.com/x0k/ps2-spy/internal/stats_tracker/storage_tasks_repo"
	stats_tracker_tasks_creator "github.com/x0k/ps2-spy/internal/stats_tracker/tasks_creator"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"github.com/x0k/ps2-spy/internal/tracking"
	tracking_settings_data_loader "github.com/x0k/ps2-spy/internal/tracking/settings_data_loader"
	tracking_settings_diff_view_loader "github.com/x0k/ps2-spy/internal/tracking/settings_diff_view_loader"
	tracking_settings_updater "github.com/x0k/ps2-spy/internal/tracking/settings_updater"
	tracking_settings_view_loader "github.com/x0k/ps2-spy/internal/tracking/settings_view_loader"
	tracking_storage_settings_repo "github.com/x0k/ps2-spy/internal/tracking/storage_settings_repo"
	tracking_storage_tracking_repo "github.com/x0k/ps2-spy/internal/tracking/storage_tracking_repo"
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

	storePubSub := pubsub.New[storage.EventType]()

	store := sql_storage.New(
		log.With(sl.Component("storage")),
		cfg.Storage.Path,
		cfg.StatsTracker.MaxTrackingDuration,
		storePubSub,
	)
	m.PreStartR("storage", store.Open)
	m.PostStopR("storage", store.Close)

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
		Transport: metrics.InstrumentTransport(
			mt,
			metrics.DefaultTransportName,
			http.DefaultTransport,
		),
	}

	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.Census.ServiceId, httpClient)

	censusDataProvider := census_data_provider.New(
		log.With(sl.Component("census_data_provider")),
		censusClient,
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
		census2.NewClient("https://census.lithafalcon.cc", cfg.Census.ServiceId, httpClient),
	)
	saerroDataProvider := saerro_data_provider.New(
		saerro.NewClient("https://saerro.ps2.live", httpClient),
	)

	facilityCache := sql_facility_cache.New(
		log.With(sl.Component("facility_cache")),
		store,
	)

	worldTrackerSubsMangers := make(map[ps2_platforms.Platform]pubsub.SubscriptionsManager[worlds_tracker.EventType], len(ps2_platforms.Platforms))
	worldTrackers := make(map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker, len(ps2_platforms.Platforms))
	charactersLoaders := make(map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms))
	characterLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.CharacterId, ps2.Character], len(ps2_platforms.Platforms))
	outfitsLoaders := make(map[ps2_platforms.Platform]loader.Multi[ps2.OutfitId, ps2.Outfit], len(ps2_platforms.Platforms))
	outfitLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit], len(ps2_platforms.Platforms))
	facilityLoaders := make(map[ps2_platforms.Platform]loader.Keyed[ps2.FacilityId, ps2.Facility], len(ps2_platforms.Platforms))

	censusCharactersRepo := ps2_census_characters_repo.New(
		log.With(sl.Component("census_characters_repo")),
		censusClient,
	)
	censusOutfitsRepo := ps2_census_outfits_repo.New(
		log.With(sl.Component("census_outfits_repo")),
		censusClient,
	)

	storageOutfitsRepo := ps2_storage_outfits_repo.New(store)
	ps2PubSub := pubsub.New[ps2.EventType]()
	outfitMembersSynchronizer := ps2_outfit_members_synchronizer.New(
		log.With(sl.Component("outfit_members_synchronizer")),
		storageOutfitsRepo,
		censusOutfitsRepo,
		cfg.Ps2.OutfitsSynchronizeInterval,
		ps2PubSub,
	)
	m.AppendVR("outfit_members_synchronizer", outfitMembersSynchronizer.Start)

	storageTrackingRepo := tracking_storage_tracking_repo.New(store)

	trackingManager := tracking.New(
		log.With(sl.Component("tracking_manager")),
		func(ctx context.Context, platform ps2_platforms.Platform, characterId ps2.CharacterId) (ps2.Character, error) {
			return characterLoaders[platform](ctx, characterId)
		},
		func(ctx context.Context, platform ps2_platforms.Platform, c ps2.Character) ([]discord.Channel, error) {
			return store.TrackingChannelsForCharacter(ctx, platform, c.Id, c.OutfitId)
		},
		store.AllTrackableCharacterIdsWithDuplicationsForPlatform,
		store.OutfitMembers,
		store.TrackingChannelsForOutfit,
		store.AllTrackableOutfitIdsWithDuplicationsForPlatform,
	)
	m.AppendVR("tracking_manager", trackingManager.Start)

	statsTrackerPubSub := pubsub.New[stats_tracker.EventType]()
	storageTasksRepo := stats_tracker_storage_tasks_repo.New(store)
	statsTracker := stats_tracker.New(
		log.With(sl.Component("stats_tracker")),
		statsTrackerPubSub,
		storageTrackingRepo.PlatformsByChannelId,
		storageTasksRepo.ChannelsWithActiveTasks,
		func(ctx context.Context, platform ps2_platforms.Platform, charId ps2.CharacterId) ([]discord.ChannelId, error) {
			channels, err := trackingManager.CharacterChannels(ctx, platform, charId)
			if err != nil {
				return nil, err
			}
			channelIds := make([]discord.ChannelId, 0, len(channels))
			for _, channel := range channels {
				channelIds = append(channelIds, channel.Id)
			}
			return channelIds, nil
		},
		func(
			ctx context.Context, platform ps2_platforms.Platform, characterIds []ps2.CharacterId,
		) (map[ps2.CharacterId]ps2.Character, error) {
			return charactersLoaders[platform](ctx, characterIds)
		},
		cfg.StatsTracker.MaxTrackingDuration,
	)
	m.AppendVR("stats_tracker", statsTracker.Start)

	charactersTrackerPubSub := pubsub.New[characters_tracker.EventType]()
	charactersTracker := characters_tracker.New(
		log.With(sl.Component("platforms_characters_tracker")),
		func(ctx context.Context, platform ps2_platforms.Platform, characterId ps2.CharacterId) (ps2.Character, error) {
			return characterLoaders[platform](ctx, characterId)
		},
		charactersTrackerPubSub,
		mt,
	)
	m.AppendVR("characters_tracker", charactersTracker.Start)

	for _, platform := range ps2_platforms.Platforms {
		pl := log.With(slog.String("platform", string(platform)))
		ns := ps2_platforms.PlatformNamespace(platform)

		eventsPubSub := pubsub.New[events.EventType]()

		eventsModule, err := events_module.New(
			pl.With(sl.Module("events")),
			platform,
			cfg.Census.StreamingEndpoint,
			cfg.Census.ServiceId,
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
		m.AppendVR(
			fmt.Sprintf("%s.batched_characters_loader", platform),
			batchedCharactersLoader.Start,
		)

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

		outfitsLoader := loader.WithMultiCache(
			log.Logger.With(sl.Component("outfits_loader_cache")),
			func(ctx context.Context, k []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, error) {
				return censusDataProvider.Outfits(ctx, platform, k)
			},
			sql_outfits_cache.New(
				log.With(sl.Component("outfits_cache")),
				store,
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

	outfitMemberSaved := ps2.Subscribe[ps2.OutfitMembersAdded](m, ps2PubSub)
	outfitMemberDeleted := ps2.Subscribe[ps2.OutfitMembersRemoved](m, ps2PubSub)
	m.AppendVR("storage_events_subscription", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-outfitMemberSaved:
				trackingManager.TrackOutfitMembers(e.OutfitId, e.Platform, e.CharacterIds)
			case e := <-outfitMemberDeleted:
				trackingManager.UntrackOutfitMembers(e.OutfitId, e.Platform, e.CharacterIds)
			}
		}
	})

	populationLoaders := map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]]{
		"spy": func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
			total := 0
			worlds := make([]ps2.WorldPopulation, 0)
			for _, platform := range ps2_platforms.Platforms {
				population := charactersTracker.WorldsPopulation(platform)
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
			population, err := charactersTracker.DetailedWorldPopulation(platform, worldId)
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

	storageSettingsRepo := tracking_storage_settings_repo.New(store)
	trackingPubSub := pubsub.New[tracking.EventType]()

	settingsUpdate := tracking.Subscribe[tracking.TrackingSettingsUpdated](m, trackingPubSub)
	m.AppendVR("tracking_settings_events_subscription", func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-settingsUpdate:
				trackingManager.HandleTrackingSettingsUpdate(ctx, e.Platform, e)
				for _, oId := range e.Diff.Outfits.ToAdd {
					outfitMembersSynchronizer.SyncOutfit(ctx, e.Platform, oId)
				}
			}
		}
	})

	statsTrackerTasksCreator := stats_tracker_tasks_creator.New(
		storageTasksRepo,
		cfg.StatsTracker.MaxTrackingDuration,
		cfg.StatsTracker.MaxNumberOfTasksPerChannel,
	)

	discordMessages := discord_messages.New(
		shared.Timezones,
		cfg.StatsTracker.MaxTrackingDuration,
		cfg.Tracking.MaxNumberTrackedCharacters,
		cfg.Tracking.MaxNumberTrackedOutfits,
		cfg.StatsTracker.MaxNumberOfTasksPerChannel,
	)
	discordCommands := discord_commands.New(
		log.With(sl.Component("commands")),
		discordMessages,
		populationLoaders,
		[]string{"spy", "honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
		worldPopulationLoaders,
		[]string{"spy", "honu", "saerro", "voidwell"},
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
		[]string{"spy", "ps2alerts", "honu", "census", "voidwell"},
		tracking_settings_data_loader.New(
			storageSettingsRepo,
			charactersTracker,
		).Load,
		func(
			ctx context.Context, platform ps2_platforms.Platform, outfitIds []ps2.OutfitId,
		) (map[ps2.OutfitId]ps2.Outfit, error) {
			return outfitsLoaders[platform](ctx, outfitIds)
		},
		tracking_settings_view_loader.New(
			storageSettingsRepo,
			censusOutfitsRepo,
			censusCharactersRepo,
		).Load,
		tracking_settings_updater.New(
			storageSettingsRepo,
			censusOutfitsRepo,
			censusCharactersRepo,
			cfg.Tracking.MaxNumberTrackedOutfits,
			cfg.Tracking.MaxNumberTrackedCharacters,
			trackingPubSub,
		).Update,
		statsTracker,
		store.Channel,
		store.SaveChannelLanguage,
		store.SaveChannelCharacterNotifications,
		store.SaveChannelOutfitNotifications,
		store.SaveChannelTitleUpdates,
		store.SaveChannelDefaultTimezone,
		storageTasksRepo.ByChannelId,
		statsTrackerTasksCreator.Create,
		storageTasksRepo.Delete,
		storageTasksRepo.ById,
		statsTrackerTasksCreator.Update,
	)
	m.AppendR("discord.commands", discordCommands.Start)

	discordModule, err := discord_module.New(
		log.With(sl.Module("discord")),
		cfg.Discord.Token,
		cfg.Discord.CommandHandlerTimeout,
		cfg.Discord.EventHandlerTimeout,
		cfg.Discord.RemoveCommands,
		discordMessages,
		discordCommands,
		trackingManager,
		storePubSub,
		ps2PubSub,
		trackingPubSub,
		charactersTrackerPubSub,
		worldTrackerSubsMangers,
		characterLoaders,
		outfitLoaders,
		charactersLoaders,
		facilityLoaders,
		func(ctx context.Context, channelId discord.ChannelId) (int, error) {
			count := 0
			errs := make([]error, 0, len(ps2_platforms.Platforms))
			for _, platform := range ps2_platforms.Platforms {
				settings, err := storageSettingsRepo.Get(ctx, channelId, platform)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				outfits, err := charactersTracker.OnlineOutfitMembers(ctx, platform, settings.Outfits)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				for _, outfit := range outfits {
					count += len(outfit)
				}
				characters, err := charactersTracker.OnlineCharacters(ctx, platform, settings.Characters)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				count += len(characters)
			}
			return count, errors.Join(errs...)
		},
		statsTrackerPubSub,
		store.Channel,
		tracking_settings_diff_view_loader.New(
			censusOutfitsRepo,
			censusCharactersRepo,
		).Load,
	)
	if err != nil {
		return nil, err
	}
	m.Append(discordModule)

	return m, nil
}
