package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/submit/channel_setup_submit_handler"
	"github.com/x0k/ps2-spy/internal/cache/facility_cache"
	"github.com/x0k/ps2-spy/internal/cache/outfit_cache"
	"github.com/x0k/ps2-spy/internal/cache/outfits_cache"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/loaders/alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/characters_loader"
	"github.com/x0k/ps2-spy/internal/loaders/event_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/facility_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_tags_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_outfit_tags_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_outfits_loader"
	"github.com/x0k/ps2-spy/internal/loaders/population_loader"
	"github.com/x0k/ps2-spy/internal/loaders/subscription_settings_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_online_entities_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_population_loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/savers/subscription_settings_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func start(ctx context.Context, log *logger.Logger, cfg *config.Config) error {
	const op = "start"
	storageEventsPublisher := publisher.New(storage.CastHandler)
	sqlStorage, err := startStorage(ctx, log, cfg.Storage, storageEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	All := []string{"all"}
	EventName := []string{
		ps2events.PlayerLoginEventName,
		ps2events.PlayerLogoutEventName,
		ps2events.AchievementEarnedEventName,
		ps2events.BattleRankUpEventName,
		ps2events.DeathEventName,
		ps2events.GainExperienceEventName,
		ps2events.ItemAddedEventName,
		ps2events.PlayerFacilityCaptureEventName,
		ps2events.PlayerFacilityDefendEventName,
		ps2events.SkillAddedEventName,
		ps2events.VehicleDestroyEventName,
		ps2events.FacilityControlEventName,
		ps2events.MetagameEventEventName,
		ps2events.ContinentLockEventName,
	}
	pcPs2EventsPublisher, err := startNewPs2EventsPublisher(ctx, log, cfg, platforms.PC, ps2commands.SubscriptionSettings{
		Worlds:     All,
		Characters: All,
		EventNames: EventName,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4euPs2EventsPublisher, err := startNewPs2EventsPublisher(ctx, log, cfg, platforms.PS4_EU, ps2commands.SubscriptionSettings{
		Worlds:     All,
		Characters: All,
		EventNames: EventName,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4usPs2EventsPublisher, err := startNewPs2EventsPublisher(ctx, log, cfg, platforms.PS4_US, ps2commands.SubscriptionSettings{
		Worlds:     All,
		Characters: All,
		EventNames: EventName,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	httpClient := &http.Client{
		Timeout: cfg.HttpClientTimeout,
	}
	// loaders
	wg := infra.Wg(ctx)
	honuClient := honu.NewClient("https://wt.honu.pw", httpClient)
	honuClient.Start(ctx, wg)
	fisuClient := fisu.NewClient("https://ps2.fisu.pw", httpClient)
	fisuClient.Start(ctx, wg)
	voidWellClient := voidwell.NewClient("https://api.voidwell.com", httpClient)
	voidWellClient.Start(ctx, wg)
	populationClient := population.NewClient("https://agg.ps2.live", httpClient)
	populationClient.Start(ctx, wg)
	saerroClient := saerro.NewClient("https://saerro.ps2.live", httpClient)
	saerroClient.Start(ctx, wg)
	ps2alertsClient := ps2alerts.NewClient("https://api.ps2alerts.com", httpClient)
	ps2alertsClient.Start(ctx, wg)
	censusClient := census2.NewClient(log.Logger, "https://census.daybreakgames.com", cfg.CensusServiceId, httpClient)
	sanctuaryClient := census2.NewClient(log.Logger, "https://census.lithafalcon.cc", cfg.CensusServiceId, httpClient)

	// TODO: Use MultiKeyedCache
	pcCharactersLoader := characters_loader.NewCensus(log, censusClient, platforms.PC)
	pcBatchedCharacterLoader := character_loader.NewBatch(log.With(slog.String("platform", string(platforms.PC))), pcCharactersLoader, 10*time.Second)
	pcBatchedCharacterLoader.Start(ctx, wg)

	ps4euCharactersLoader := characters_loader.NewCensus(log, censusClient, platforms.PS4_EU)
	ps4euBatchedCharacterLoader := character_loader.NewBatch(log.With(slog.String("platform", string(platforms.PS4_EU))), ps4euCharactersLoader, 10*time.Second)
	ps4euBatchedCharacterLoader.Start(ctx, wg)

	ps4usCharactersLoader := characters_loader.NewCensus(log, censusClient, platforms.PS4_US)
	ps4usBatchedCharacterLoader := character_loader.NewBatch(log.With(slog.String("platform", string(platforms.PS4_US))), ps4usCharactersLoader, 10*time.Second)
	ps4usBatchedCharacterLoader.Start(ctx, wg)

	pcCharactersTrackerPublisher := publisher.New(characters_tracker.CastHandler)
	pcCharactersTracker, err := startNewCharactersTracker(
		ctx,
		log,
		ps2.PcPlatformWorldIds,
		pcBatchedCharacterLoader,
		pcPs2EventsPublisher,
		pcCharactersTrackerPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ps4euCharactersTrackerPublisher := publisher.New(characters_tracker.CastHandler)
	ps4euCharactersTracker, err := startNewCharactersTracker(
		ctx,
		log,
		ps2.Ps4euPlatformWorldIds,
		ps4euBatchedCharacterLoader,
		ps4euPs2EventsPublisher,
		ps4euCharactersTrackerPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ps4usCharactersTrackerPublisher := publisher.New(characters_tracker.CastHandler)
	ps4usCharactersTracker, err := startNewCharactersTracker(
		ctx,
		log,
		ps2.Ps4usPlatformWorldIds,
		ps4usBatchedCharacterLoader,
		ps4usPs2EventsPublisher,
		ps4usCharactersTrackerPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	platformCharactersTrackers := map[platforms.Platform]*characters_tracker.CharactersTracker{
		platforms.PC:     pcCharactersTracker,
		platforms.PS4_EU: ps4euCharactersTracker,
		platforms.PS4_US: ps4usCharactersTracker,
	}

	pcWorldsTrackerPublisher := publisher.New(worlds_tracker.CastHandler)
	pcWorldsTracker, err := startNewWorldsTracker(ctx, log, pcPs2EventsPublisher, pcWorldsTrackerPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ps4euWorldsTrackerPublisher := publisher.New(worlds_tracker.CastHandler)
	ps4euWorldsTracker, err := startNewWorldsTracker(ctx, log, ps4euPs2EventsPublisher, ps4euWorldsTrackerPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ps4usWorldsTrackerPublisher := publisher.New(worlds_tracker.CastHandler)
	ps4usWorldsTracker, err := startNewWorldsTracker(ctx, log, ps4usPs2EventsPublisher, ps4usWorldsTrackerPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	platformWorldsTrackers := map[platforms.Platform]*worlds_tracker.WorldsTracker{
		platforms.PC:     pcWorldsTracker,
		platforms.PS4_EU: ps4euWorldsTracker,
		platforms.PS4_US: ps4usWorldsTracker,
	}

	// multi loaders
	popLoader := population_loader.NewMulti(
		log,
		map[string]loaders.Loader[loaders.Loaded[ps2.WorldsPopulation]]{
			// TODO: Add tiny cache for spy loaders
			"spy": population_loader.NewCharactersTrackerLoader(
				log,
				cfg.BotName,
				platformCharactersTrackers,
			),
			"honu":      population_loader.NewHonu(honuClient),
			"ps2live":   population_loader.NewPS2Live(populationClient),
			"saerro":    population_loader.NewSaerro(saerroClient),
			"fisu":      population_loader.NewFisu(fisuClient),
			"sanctuary": population_loader.NewSanctuary(sanctuaryClient),
			"voidwell":  population_loader.NewVoidWell(voidWellClient),
		},
		[]string{"spy", "honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
	)
	popLoader.Start(ctx, wg)
	worldPopLoader := world_population_loader.NewMulti(
		log,
		map[string]loaders.KeyedLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]]{
			"spy": world_population_loader.NewCharactersTrackerLoader(
				cfg.BotName,
				platformCharactersTrackers,
			),
			"honu":     world_population_loader.NewHonu(honuClient),
			"saerro":   world_population_loader.NewSaerro(saerroClient),
			"voidwell": world_population_loader.NewVoidWell(voidWellClient),
		},
		[]string{"spy", "honu", "saerro", "voidwell"},
	)
	worldPopLoader.Start(ctx, wg)
	alertsLoader := alerts_loader.NewMulti(
		log,
		map[string]loaders.Loader[loaders.Loaded[ps2.Alerts]]{
			"spy":       alerts_loader.NewWorldsTrackerLoader(log, cfg.BotName, platformWorldsTrackers),
			"ps2alerts": alerts_loader.NewPS2Alerts(ps2alertsClient),
			"honu":      alerts_loader.NewHonu(honuClient),
			"census":    alerts_loader.NewCensus(log, censusClient),
			"voidwell":  alerts_loader.NewVoidWell(voidWellClient),
		},
		[]string{"spy", "ps2alerts", "honu", "census", "voidwell"},
	)
	alertsLoader.Start(ctx, wg)
	worldAlertsLoader := world_alerts_loader.NewMulti(alertsLoader)
	worldAlertsLoader.Start(ctx, wg)

	characterTrackingChannelsLoader := character_tracking_channels_loader.New(sqlStorage)
	pcTrackingManager := newTrackingManager(
		log,
		sqlStorage,
		pcBatchedCharacterLoader,
		characterTrackingChannelsLoader,
		platforms.PC,
	)
	ps4euTrackingManager := newTrackingManager(
		log,
		sqlStorage,
		ps4euBatchedCharacterLoader,
		characterTrackingChannelsLoader,
		platforms.PS4_EU,
	)
	ps4usTrackingManager := newTrackingManager(
		log,
		sqlStorage,
		ps4usBatchedCharacterLoader,
		characterTrackingChannelsLoader,
		platforms.PS4_US,
	)
	err = startTrackingManager(
		ctx,
		log,
		map[platforms.Platform]*tracking_manager.TrackingManager{
			platforms.PC:     pcTrackingManager,
			platforms.PS4_EU: ps4euTrackingManager,
			platforms.PS4_US: ps4usTrackingManager,
		},
		storageEventsPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	pcOutfitMembersSaverPublisher := publisher.New(outfit_members_saver.CastHandler)
	ps4euOutfitMembersSaverPublisher := publisher.New(outfit_members_saver.CastHandler)
	ps4usOutfitMembersSaverPublisher := publisher.New(outfit_members_saver.CastHandler)
	err = startOutfitMembersSynchronizers(
		ctx,
		log,
		sqlStorage,
		censusClient,
		storageEventsPublisher,
		map[platforms.Platform]publisher.Abstract[publisher.Event]{
			platforms.PC:     pcOutfitMembersSaverPublisher,
			platforms.PS4_EU: ps4euOutfitMembersSaverPublisher,
			platforms.PS4_US: ps4usOutfitMembersSaverPublisher,
		},
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	platformCharacterNamesLoader := platform_character_names_loader.NewCensus(censusClient)
	platformOutfitTagsLoader := platform_outfit_tags_loader.NewCensus(censusClient)
	subSettingsLoader := subscription_settings_loader.New(sqlStorage)

	pcOutfitLoader := loaders.NewCachedQueriedLoader(
		outfit_loader.NewCensus(censusClient, platforms.PC),
		outfit_cache.NewStorage(log, sqlStorage, platforms.PC),
	)
	ps4euOutfitLoader := loaders.NewCachedQueriedLoader(
		outfit_loader.NewCensus(censusClient, platforms.PS4_EU),
		outfit_cache.NewStorage(log, sqlStorage, platforms.PS4_EU),
	)
	ps4usOutfitLoader := loaders.NewCachedQueriedLoader(
		outfit_loader.NewCensus(censusClient, platforms.PS4_US),
		outfit_cache.NewStorage(log, sqlStorage, platforms.PS4_US),
	)

	facilityCache := facility_cache.NewStorage(log, sqlStorage)
	pcFacilityLoader := loaders.NewCachedQueriedLoader(
		facility_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
		facilityCache,
	)
	ps4euFacilityLoader := loaders.NewCachedQueriedLoader(
		facility_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
		facilityCache,
	)
	ps4usFacilityLoader := loaders.NewCachedQueriedLoader(
		facility_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
		facilityCache,
	)

	// bot
	botConfig := &bot.BotConfig{
		DiscordToken:           cfg.DiscordToken,
		RemoveCommands:         cfg.RemoveCommands,
		CommandHandlerTimeout:  cfg.CommandHandlerTimeout,
		Ps2EventHandlerTimeout: cfg.Ps2EventHandlerTimeout,
		Commands: bot.NewCommands(
			popLoader,
			worldPopLoader,
			alertsLoader,
		),
		CommandHandlers: bot.NewCommandHandlers(
			log,
			popLoader,
			worldPopLoader,
			alertsLoader,
			worldAlertsLoader,
			subSettingsLoader,
			platformCharacterNamesLoader,
			platformOutfitTagsLoader,
			trackable_online_entities_loader.New(
				subSettingsLoader,
				map[platforms.Platform]*characters_tracker.CharactersTracker{
					platforms.PC:     pcCharactersTracker,
					platforms.PS4_EU: ps4euCharactersTracker,
					platforms.PS4_US: ps4usCharactersTracker,
				},
			),
			loaders.NewCachedQueriedLoader(
				platform_outfits_loader.NewCensus(censusClient),
				meta.NewPlatformsCache(map[platforms.Platform]containers.MultiKeyedCache[ps2.OutfitId, ps2.Outfit]{
					platforms.PC:     outfits_cache.NewStorage(log, sqlStorage, platforms.PC),
					platforms.PS4_EU: outfits_cache.NewStorage(log, sqlStorage, platforms.PS4_EU),
					platforms.PS4_US: outfits_cache.NewStorage(log, sqlStorage, platforms.PS4_US),
				}),
			),
		),
		SubmitHandlers: map[string]handlers.InteractionHandler{
			handlers.CHANNEL_SETUP_PC_MODAL: channel_setup_submit_handler.New(
				character_ids_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				character_names_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				outfit_ids_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				outfit_tags_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PC),
			),
			handlers.CHANNEL_SETUP_PS4_EU_MODAL: channel_setup_submit_handler.New(
				character_ids_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				character_names_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				outfit_ids_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				outfit_tags_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PS4_EU),
			),
			handlers.CHANNEL_SETUP_PS4_US_MODAL: channel_setup_submit_handler.New(
				character_ids_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				character_names_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				outfit_ids_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				outfit_tags_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PS4_US),
			),
		},
	}
	return startNewBot(
		ctx,
		log,
		botConfig,
		event_tracking_channels_loader.New(pcTrackingManager),
		bot.NewEventHandlers(
			log,
			pcCharactersTrackerPublisher,
			pcOutfitMembersSaverPublisher,
			pcWorldsTrackerPublisher,
			pcBatchedCharacterLoader,
			pcOutfitLoader,
			pcFacilityLoader,
			pcCharactersLoader,
		),
		event_tracking_channels_loader.New(ps4euTrackingManager),
		bot.NewEventHandlers(
			log,
			ps4euCharactersTrackerPublisher,
			ps4euOutfitMembersSaverPublisher,
			ps4euWorldsTrackerPublisher,
			ps4euBatchedCharacterLoader,
			ps4euOutfitLoader,
			ps4euFacilityLoader,
			ps4euCharactersLoader,
		),
		event_tracking_channels_loader.New(ps4usTrackingManager),
		bot.NewEventHandlers(
			log,
			ps4usCharactersTrackerPublisher,
			ps4usOutfitMembersSaverPublisher,
			ps4usWorldsTrackerPublisher,
			ps4usBatchedCharacterLoader,
			ps4usOutfitLoader,
			ps4usFacilityLoader,
			ps4usCharactersLoader,
		),
	)
}
