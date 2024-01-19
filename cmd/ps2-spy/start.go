package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/login"
	"github.com/x0k/ps2-spy/internal/bot/handlers/submit/channel_setup_submit_handler"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
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
	"github.com/x0k/ps2-spy/internal/loaders/platform_character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/platform_outfit_tags_loader"
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
	log := infra.OpLogger(ctx, op)
	storageEventsPublisher := storage.NewPublisher()
	sqlStorage, err := startStorage(ctx, cfg.Storage, storageEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	pcStreamingClient := streaming.NewClient(
		log,
		"wss://push.planetside2.com/streaming",
		streaming.Ps2_env,
		cfg.CensusServiceId,
	)
	startStreamingClient(ctx, cfg, pcStreamingClient, ps2commands.SubscriptionSettings{
		Worlds: []string{"1", "10", "13", "17", "19", "40"},
		EventNames: []string{
			ps2events.PlayerLoginEventName,
			ps2events.PlayerLogoutEventName,
		},
	})
	pcEventsPublisher := ps2events.NewPublisher()
	err = startPs2EventsPublisher(ctx, cfg, pcStreamingClient.Msg, pcEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4euStreamingClient := streaming.NewClient(
		log,
		"wss://push.planetside2.com/streaming",
		streaming.Ps2ps4eu_env,
		cfg.CensusServiceId,
	)
	startStreamingClient(ctx, cfg, ps4euStreamingClient, ps2commands.SubscriptionSettings{
		Worlds: []string{"2000"},
		EventNames: []string{
			ps2events.PlayerLoginEventName,
			ps2events.PlayerLogoutEventName,
		},
	})
	ps4euEventsPublisher := ps2events.NewPublisher()
	err = startPs2EventsPublisher(ctx, cfg, ps4euStreamingClient.Msg, ps4euEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4usStreamingClient := streaming.NewClient(
		log,
		"wss://push.planetside2.com/streaming",
		streaming.Ps2ps4us_env,
		cfg.CensusServiceId,
	)
	startStreamingClient(ctx, cfg, ps4usStreamingClient, ps2commands.SubscriptionSettings{
		Worlds: []string{"1000"},
		EventNames: []string{
			ps2events.PlayerLoginEventName,
			ps2events.PlayerLogoutEventName,
		},
	})
	ps4usEventsPublisher := ps2events.NewPublisher()
	err = startPs2EventsPublisher(ctx, cfg, ps4usStreamingClient.Msg, ps4usEventsPublisher)
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
	// TODO: remove loader
	err = startOutfitMembersSynchronizer(
		ctx,
		sqlStorage,
		censusClient,
		storageEventsPublisher,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	platformCharacterNamesLoader := platform_character_names_loader.NewCensus(censusClient)
	platformOutfitTagsLoader := platform_outfit_tags_loader.NewCensus(censusClient)

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
			platformCharacterNamesLoader,
			platformOutfitTagsLoader,
		),
		SubmitHandlers: map[string]handlers.InteractionHandler{
			handlers.CHANNEL_SETUP_PC_MODAL: channel_setup_submit_handler.New(
				character_ids_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				character_names_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				outfit_tags_loader.NewCensus(censusClient, census2.Ps2_v2_NS),
				subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PC),
			),
			handlers.CHANNEL_SETUP_PS4_EU_MODAL: channel_setup_submit_handler.New(
				character_ids_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				character_names_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				outfit_tags_loader.NewCensus(censusClient, census2.Ps2ps4eu_v2_NS),
				subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PS4_EU),
			),
			handlers.CHANNEL_SETUP_PS4_US_MODAL: channel_setup_submit_handler.New(
				character_ids_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				character_names_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				outfit_tags_loader.NewCensus(censusClient, census2.Ps2ps4us_v2_NS),
				subscription_settings_saver.New(sqlStorage, subSettingsLoader, platforms.PS4_US),
			),
		},
		PlayerLoginHandlers: map[string]handlers.Ps2EventHandler[ps2events.PlayerLogin]{
			platforms.PC:     login.New(pcBatchedCharacterLoader),
			platforms.PS4_EU: login.New(ps4euBatchedCharacterLoader),
			platforms.PS4_US: login.New(ps4usBatchedCharacterLoader),
		},
		EventsPublishers: map[string]*ps2events.Publisher{
			platforms.PC:     pcEventsPublisher,
			platforms.PS4_EU: ps4euEventsPublisher,
			platforms.PS4_US: ps4usEventsPublisher,
		},
		EventTrackingChannelsLoaders: map[string]loaders.QueriedLoader[any, []string]{
			platforms.PC:     event_tracking_channels_loader.New(pcTrackingManager),
			platforms.PS4_EU: event_tracking_channels_loader.New(ps4euTrackingManager),
			platforms.PS4_US: event_tracking_channels_loader.New(ps4usTrackingManager),
		},
	}
	return startBot(ctx, botConfig)
}
