package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/login"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/characters_loader"
	"github.com/x0k/ps2-spy/internal/loaders/event_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_tags_loader"
	"github.com/x0k/ps2-spy/internal/loaders/population_loader"
	"github.com/x0k/ps2-spy/internal/loaders/subscription_settings_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_population_loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/subscription_settings_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func start(ctx context.Context, cfg *config.Config) error {
	const op = "start"
	storageEventsPublisher := storage.NewPublisher()
	sqlStorage, err := startStorage(ctx, cfg.Storage, storageEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	eventsPublisher, err := startPs2EventsPublisher(ctx, cfg)
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
	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.CensusServiceId, httpClient)
	sanctuaryClient := census2.NewClient("https://census.lithafalcon.cc", cfg.CensusServiceId, httpClient)
	// multi loaders
	log := infra.OpLogger(ctx, op)
	popLoader := population_loader.NewMulti(
		log,
		map[string]loaders.Loader[loaders.Loaded[ps2.WorldsPopulation]]{
			"honu":      population_loader.NewHonu(honuClient),
			"ps2live":   population_loader.NewPS2Live(populationClient),
			"saerro":    population_loader.NewSaerro(saerroClient),
			"fisu":      population_loader.NewFisu(fisuClient),
			"sanctuary": population_loader.NewSanctuary(sanctuaryClient),
			"voidwell":  population_loader.NewVoidWell(voidWellClient),
		},
		[]string{"honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
	)
	popLoader.Start(ctx, wg)
	worldPopLoader := world_population_loader.NewMulti(
		log,
		map[string]loaders.KeyedLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]]{
			"honu":     world_population_loader.NewHonu(honuClient),
			"saerro":   world_population_loader.NewSaerro(saerroClient),
			"voidwell": world_population_loader.NewVoidWell(voidWellClient),
		},
		[]string{"honu", "saerro", "voidwell"},
	)
	worldPopLoader.Start(ctx, wg)
	alertsLoader := alerts_loader.NewMulti(
		log,
		map[string]loaders.Loader[loaders.Loaded[ps2.Alerts]]{
			"ps2alerts": alerts_loader.NewPS2Alerts(ps2alertsClient),
			"honu":      alerts_loader.NewHonu(honuClient),
			"census":    alerts_loader.NewCensus(censusClient),
			"voidwell":  alerts_loader.NewVoidWell(voidWellClient),
		},
		[]string{"ps2alerts", "honu", "census", "voidwell"},
	)
	alertsLoader.Start(ctx, wg)
	worldAlertsLoader := world_alerts_loader.NewMulti(alertsLoader)
	worldAlertsLoader.Start(ctx, wg)

	pcCharactersLoader := characters_loader.NewCensus(censusClient, census2.Ps2_v2_NS)
	pcBatchedCharacterLoader := character_loader.NewBatch(pcCharactersLoader, time.Minute)
	pcBatchedCharacterLoader.Start(ctx, wg)

	ps4euCharactersLoader := characters_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS)
	ps4euBatchedCharacterLoader := character_loader.NewBatch(ps4euCharactersLoader, time.Minute)
	ps4euBatchedCharacterLoader.Start(ctx, wg)

	ps4usCharactersLoader := characters_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS)
	ps4usBatchedCharacterLoader := character_loader.NewBatch(ps4usCharactersLoader, time.Minute)
	ps4usBatchedCharacterLoader.Start(ctx, wg)

	characterTrackingChannelsLoader := character_tracking_channels_loader.New(sqlStorage)
	pcTrackingManager := newTrackingManager(
		sqlStorage,
		pcBatchedCharacterLoader,
		characterTrackingChannelsLoader,
		platforms.PC,
	)
	ps4euTrackingManager := newTrackingManager(
		sqlStorage,
		ps4euBatchedCharacterLoader,
		characterTrackingChannelsLoader,
		platforms.PS4_EU,
	)
	ps4usTrackingManager := newTrackingManager(
		sqlStorage,
		ps4usBatchedCharacterLoader,
		characterTrackingChannelsLoader,
		platforms.PS4_US,
	)
	err = startTrackingManager(
		ctx,
		map[string]*tracking_manager.TrackingManager{
			platforms.PC:     pcTrackingManager,
			platforms.PS4_EU: ps4euTrackingManager,
			platforms.PS4_US: ps4usTrackingManager,
		},
		storageEventsPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	subSettingsLoader := subscription_settings_loader.New(sqlStorage)
	pcCharacterNamesLoader := character_names_loader.NewCensus(censusClient, census2.Ps2_v2_NS)
	pcOutfitTagsLoader := outfit_tags_loader.NewCensus(censusClient, census2.Ps2_v2_NS)
	err = startOutfitMembersSynchronizer(
		ctx,
		sqlStorage,
		censusClient,
		storageEventsPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// bot
	botConfig := &bot.BotConfig{
		DiscordToken:           cfg.DiscordToken,
		CommandHandlerTimeout:  cfg.CommandHandlerTimeout,
		Ps2EventHandlerTimeout: cfg.Ps2EventHandlerTimeout,
		Commands: bot.NewCommands(
			popLoader,
			worldPopLoader,
			alertsLoader,
		),
		CommandHandlers: bot.NewCommandHandlers(
			popLoader,
			worldPopLoader,
			alertsLoader,
			worldAlertsLoader,
			subSettingsLoader,
			pcCharacterNamesLoader,
			pcOutfitTagsLoader,
		),
		SubmitHandlers: bot.NewSubmitHandlers(
			character_ids_loader.NewCensus(censusClient),
			pcCharacterNamesLoader,
			pcOutfitTagsLoader,
			subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PC),
			subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PS4_EU),
			subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PS4_US),
		),
		EventsPublisher:             eventsPublisher,
		PlayerLoginHandler:          login.New(pcBatchedCharacterLoader),
		EventTrackingChannelsLoader: event_tracking_channels_loader.New(pcTrackingManager),
	}
	return startBot(ctx, botConfig)
}
