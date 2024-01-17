package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/bot/handlers/event/login"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_loader"
	"github.com/x0k/ps2-spy/internal/loaders/character_names_loader"
	"github.com/x0k/ps2-spy/internal/loaders/characters_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_tags_loader"
	"github.com/x0k/ps2-spy/internal/loaders/population_loader"
	"github.com/x0k/ps2-spy/internal/loaders/subscription_settings_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_loader"
	"github.com/x0k/ps2-spy/internal/loaders/tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_alerts_loader"
	"github.com/x0k/ps2-spy/internal/loaders/world_population_loader"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/savers/subscription_settings_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

type setup struct {
	log *slog.Logger
	ctx context.Context
	wg  *sync.WaitGroup
}

func startBot(s *setup, cfg *config.Config) error {
	const op = "startBot"
	storageEventsPublisher := storage.NewPublisher(s.log)
	storage, err := startStorage(s, cfg.Storage, storageEventsPublisher)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	eventsPublisher, err := startPs2EventsPublisher(s, cfg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	httpClient := &http.Client{
		Timeout: cfg.HttpClientTimeout,
	}
	// loaders
	honuClient := honu.NewClient("https://wt.honu.pw", httpClient)
	honuClient.Start(s.ctx, s.wg)
	fisuClient := fisu.NewClient("https://ps2.fisu.pw", httpClient)
	fisuClient.Start(s.ctx, s.wg)
	voidWellClient := voidwell.NewClient("https://api.voidwell.com", httpClient)
	voidWellClient.Start(s.ctx, s.wg)
	populationClient := population.NewClient("https://agg.ps2.live", httpClient)
	populationClient.Start(s.ctx, s.wg)
	saerroClient := saerro.NewClient("https://saerro.ps2.live", httpClient)
	saerroClient.Start(s.ctx, s.wg)
	ps2alertsClient := ps2alerts.NewClient("https://api.ps2alerts.com", httpClient)
	ps2alertsClient.Start(s.ctx, s.wg)
	censusClient := census2.NewClient("https://census.daybreakgames.com", cfg.CensusServiceId, httpClient)
	sanctuaryClient := census2.NewClient("https://census.lithafalcon.cc", cfg.CensusServiceId, httpClient)
	// multi loaders
	popLoader := population_loader.NewMulti(
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
	popLoader.Start(s.ctx, s.wg)
	worldPopLoader := world_population_loader.NewMulti(
		map[string]loaders.KeyedLoader[ps2.WorldId, loaders.Loaded[ps2.DetailedWorldPopulation]]{
			"honu":     world_population_loader.NewHonu(honuClient),
			"saerro":   world_population_loader.NewSaerro(saerroClient),
			"voidwell": world_population_loader.NewVoidWell(voidWellClient),
		},
		[]string{"honu", "saerro", "voidwell"},
	)
	worldPopLoader.Start(s.ctx, s.wg)
	alertsLoader := alerts_loader.NewMulti(
		map[string]loaders.Loader[loaders.Loaded[ps2.Alerts]]{
			"ps2alerts": alerts_loader.NewPS2Alerts(ps2alertsClient),
			"honu":      alerts_loader.NewHonu(honuClient),
			"census":    alerts_loader.NewCensus(censusClient),
			"voidwell":  alerts_loader.NewVoidWell(voidWellClient),
		},
		[]string{"ps2alerts", "honu", "census", "voidwell"},
	)
	alertsLoader.Start(s.ctx, s.wg)
	worldAlertsLoader := world_alerts_loader.NewMulti(alertsLoader)
	worldAlertsLoader.Start(s.ctx, s.wg)
	batchedCharacterLoader := character_loader.NewBatch(s.log, characters_loader.NewCensus(censusClient))
	batchedCharacterLoader.Start(s.ctx, s.wg)
	channelsLoader := tracking_channels_loader.New(storage)
	trackingManager := tracking_manager.New(batchedCharacterLoader, channelsLoader)
	subSettingsLoader := subscription_settings_loader.New(storage)
	characterNamesLoader := character_names_loader.NewCensus(censusClient)
	outfitTagsLoader := outfit_tags_loader.NewCensus(censusClient)
	pcTrackableOutfitsLoader := trackable_outfits_loader.NewStorage(
		storage,
		platforms.PC,
	)
	outfitMembersLoader := outfit_member_ids_loader.NewCensus(censusClient)
	pcOutfitMembersSaver := outfit_members_saver.New(
		storage,
		platforms.PC,
	)
	pcOutfitMembersSynchronizer := outfit_members_synchronizer.New(
		s.log,
		pcTrackableOutfitsLoader,
		outfitMembersLoader,
		pcOutfitMembersSaver,
		time.Hour*24,
	)
	pcOutfitMembersSynchronizer.Start(s.ctx, s.wg)
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
			characterNamesLoader,
			outfitTagsLoader,
		),
		SubmitHandlers: bot.NewSubmitHandlers(
			character_ids_loader.NewCensus(censusClient),
			characterNamesLoader,
			outfitTagsLoader,
			subscription_settings_saver.New(storage, subSettingsLoader, platforms.PC),
			subscription_settings_saver.New(storage, subSettingsLoader, platforms.PS4_EU),
			subscription_settings_saver.New(storage, subSettingsLoader, platforms.PS4_US),
		),
		EventsPublisher:    eventsPublisher,
		PlayerLoginHandler: login.New(batchedCharacterLoader),
		TrackingManager:    trackingManager,
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		retry.RetryWhileWithRecover(retry.Retryable{
			Try: func() error {
				b, err := bot.New(s.ctx, s.log, botConfig)
				if err != nil {
					return err
				}
				defer func() {
					if err := b.Stop(); err != nil {
						s.log.Error("failed to stop bot", sl.Err(err))
					}
				}()
				<-s.ctx.Done()
				return s.ctx.Err()
			},
			While: retry.ContextIsNotCanceled,
			BeforeSleep: func(d time.Duration) {
				s.log.Debug("retry to start bot", slog.Duration("after", d))
			},
		})
	}()
	return nil
}
