package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/cache/facility_cache"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
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
	"github.com/x0k/ps2-spy/internal/loaders/character_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/event_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/population_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_map_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_population_loader"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func start(
	ctx context.Context,
	wg *sync.WaitGroup,
	log *logger.Logger,
	cfg *config.Config,
) error {
	const op = "start"
	startProfiler(ctx, wg, log, cfg.Profiler)
	mt := startMetrics(ctx, wg, log, cfg.Metrics)
	storageEventsPublisher := storage.NewPublisher(
		mt.InstrumentPublisher(
			metrics.StoragePublisher,
			publisher.New[publisher.Event](),
		),
	)
	sqlStorage, err := startStorage(ctx, wg, log, cfg.Storage, storageEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	subscriptionSettings := ps2commands.SubscriptionSettings{
		Worlds:     []string{"all"},
		Characters: []string{"all"},
		EventNames: []string{
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
		},
	}
	pcPs2EventsPublisher, err := startNewPs2EventsPublisher(
		ctx, log, mt, cfg, platforms.PC, subscriptionSettings)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4euPs2EventsPublisher, err := startNewPs2EventsPublisher(
		ctx, log, mt, cfg, platforms.PS4_EU, subscriptionSettings)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4usPs2EventsPublisher, err := startNewPs2EventsPublisher(
		ctx, log, mt, cfg, platforms.PS4_US, subscriptionSettings)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	httpClient := &http.Client{
		Timeout: cfg.HttpClientTimeout,
		Transport: mt.InstrumentTransport(
			metrics.DefaultTransportName,
			http.DefaultTransport,
		),
	}
	// loaders
	// TODO: extract cache from clients
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

	// TODO: Lookup in storage (stale data) after fail?
	pcCharactersLoader := newCharactersLoader(
		log, mt, platforms.PC,
		censusClient,
	)
	pcBatchedCharacterLoader := startNewBatchedCharacterLoader(
		ctx, wg, log, mt, platforms.PC,
		pcCharactersLoader,
	)

	ps4euCharactersLoader := newCharactersLoader(
		log, mt, platforms.PS4_EU,
		censusClient,
	)
	ps4euBatchedCharacterLoader := startNewBatchedCharacterLoader(
		ctx, wg, log, mt, platforms.PS4_EU,
		ps4euCharactersLoader,
	)

	ps4usCharactersLoader := newCharactersLoader(
		log, mt, platforms.PS4_US,
		censusClient,
	)
	ps4usBatchedCharacterLoader := startNewBatchedCharacterLoader(
		ctx, wg, log, mt, platforms.PS4_US,
		ps4usCharactersLoader,
	)

	pcCharactersTrackerPublisher := characters_tracker.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.CharactersTrackerPlatformPublisher,
			platforms.PC,
			publisher.New[publisher.Event](),
		),
	)
	ps4euCharactersTrackerPublisher := characters_tracker.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.CharactersTrackerPlatformPublisher,
			platforms.PS4_EU,
			publisher.New[publisher.Event](),
		),
	)
	ps4usCharactersTrackerPublisher := characters_tracker.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.CharactersTrackerPlatformPublisher,
			platforms.PS4_US,
			publisher.New[publisher.Event](),
		),
	)

	platformCharactersTrackers := map[platforms.Platform]*characters_tracker.CharactersTracker{
		platforms.PC: startNewCharactersTracker(
			ctx, log, mt, platforms.PC,
			ps2.PcPlatformWorldIds,
			pcBatchedCharacterLoader,
			pcPs2EventsPublisher,
			pcCharactersTrackerPublisher,
		),
		platforms.PS4_EU: startNewCharactersTracker(
			ctx, log, mt, platforms.PS4_EU,
			ps2.Ps4euPlatformWorldIds,
			ps4euBatchedCharacterLoader,
			ps4euPs2EventsPublisher,
			ps4euCharactersTrackerPublisher,
		),
		platforms.PS4_US: startNewCharactersTracker(
			ctx, log, mt, platforms.PS4_US,
			ps2.Ps4usPlatformWorldIds,
			ps4usBatchedCharacterLoader,
			ps4usPs2EventsPublisher,
			ps4usCharactersTrackerPublisher,
		),
	}

	pcWorldsTrackerPublisher := worlds_tracker.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.WorldsTrackerPlatformPublisher,
			platforms.PC,
			publisher.New[publisher.Event](),
		),
	)
	ps4euWorldsTrackerPublisher := worlds_tracker.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.WorldsTrackerPlatformPublisher,
			platforms.PS4_EU,
			publisher.New[publisher.Event](),
		),
	)
	ps4usWorldsTrackerPublisher := worlds_tracker.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.WorldsTrackerPlatformPublisher,
			platforms.PS4_US,
			publisher.New[publisher.Event](),
		),
	)

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
	platformWorldTrackers := map[platforms.Platform]*worlds_tracker.WorldsTracker{
		platforms.PC: startNewWorldsTracker(
			ctx, log, platforms.PC,
			pcPs2EventsPublisher,
			pcWorldsTrackerPublisher,
			world_map_loader.NewCensus(censusClient, platforms.PC),
		),
		platforms.PS4_EU: startNewWorldsTracker(
			ctx, log, platforms.PS4_EU,
			ps4euPs2EventsPublisher,
			ps4euWorldsTrackerPublisher,
			world_map_loader.NewCensus(censusClient, platforms.PS4_EU),
		),
		platforms.PS4_US: startNewWorldsTracker(
			ctx, log, platforms.PS4_US,
			ps4usPs2EventsPublisher,
			ps4usWorldsTrackerPublisher,
			world_map_loader.NewCensus(censusClient, platforms.PS4_US),
		),
	}
	alertsLoader := alerts_loader.NewMulti(
		log,
		map[string]loaders.Loader[loaders.Loaded[ps2.Alerts]]{
			"spy": alerts_loader.NewWorldsTrackerLoader(
				log,
				cfg.BotName,
				platformWorldTrackers,
			),
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
	startTrackingManager(
		ctx,
		log,
		map[platforms.Platform]*tracking_manager.TrackingManager{
			platforms.PC:     pcTrackingManager,
			platforms.PS4_EU: ps4euTrackingManager,
			platforms.PS4_US: ps4usTrackingManager,
		},
		storageEventsPublisher,
	)

	pcOutfitMembersSaverPublisher := outfit_members_saver.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.OutfitsMembersSaverPlatformPublisher,
			platforms.PC,
			publisher.New[publisher.Event](),
		),
	)
	ps4euOutfitMembersSaverPublisher := outfit_members_saver.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.OutfitsMembersSaverPlatformPublisher,
			platforms.PS4_EU,
			publisher.New[publisher.Event](),
		),
	)
	ps4usOutfitMembersSaverPublisher := outfit_members_saver.NewPublisher(
		mt.InstrumentPlatformPublisher(
			metrics.OutfitsMembersSaverPlatformPublisher,
			platforms.PS4_US,
			publisher.New[publisher.Event](),
		),
	)
	if err = startNewOutfitMembersSynchronizers(
		ctx, wg, log,
		sqlStorage,
		censusClient,
		storageEventsPublisher,
		map[platforms.Platform]*outfit_members_saver.Publisher{
			platforms.PC:     pcOutfitMembersSaverPublisher,
			platforms.PS4_EU: ps4euOutfitMembersSaverPublisher,
			platforms.PS4_US: ps4usOutfitMembersSaverPublisher,
		},
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	facilityCache := facility_cache.NewStorage(log, sqlStorage)

	// bot
	startNewBot(
		ctx,
		log,
		bot.NewConfig(
			log, cfg,
			sqlStorage,
			censusClient,
			facilityCache,
			popLoader,
			worldPopLoader,
			alertsLoader,
			worldAlertsLoader,
			platformWorldTrackers,
			platformCharactersTrackers,
		),
		event_tracking_channels_loader.New(pcTrackingManager),
		newEventHandler(
			log, mt, platforms.PC,
			pcCharactersTrackerPublisher,
			pcOutfitMembersSaverPublisher,
			pcWorldsTrackerPublisher,
			sqlStorage,
			censusClient,
			facilityCache,
			pcBatchedCharacterLoader,
			pcCharactersLoader,
		),
		event_tracking_channels_loader.New(ps4euTrackingManager),
		newEventHandler(
			log, mt, platforms.PS4_EU,
			ps4euCharactersTrackerPublisher,
			ps4euOutfitMembersSaverPublisher,
			ps4euWorldsTrackerPublisher,
			sqlStorage,
			censusClient,
			facilityCache,
			ps4euBatchedCharacterLoader,
			ps4euCharactersLoader,
		),
		event_tracking_channels_loader.New(ps4usTrackingManager),
		newEventHandler(
			log, mt, platforms.PS4_US,
			ps4usCharactersTrackerPublisher,
			ps4usOutfitMembersSaverPublisher,
			ps4usWorldsTrackerPublisher,
			sqlStorage,
			censusClient,
			facilityCache,
			ps4usBatchedCharacterLoader,
			ps4usCharactersLoader,
		),
	)
	return nil
}
