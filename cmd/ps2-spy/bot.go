package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
	ps2commands "github.com/x0k/ps2-spy/internal/lib/census2/streaming/commands"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	ps2messages "github.com/x0k/ps2-spy/internal/lib/census2/streaming/messages"
	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/ps2alerts"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/ps2/alerts"
	"github.com/x0k/ps2-spy/internal/loaders/ps2/world"
	"github.com/x0k/ps2-spy/internal/loaders/ps2/worlds"
	alertsMultiLoader "github.com/x0k/ps2-spy/internal/multi_loaders/ps2/alerts"
	popMultiLoader "github.com/x0k/ps2-spy/internal/multi_loaders/ps2/population"
	worldPopMultiLoader "github.com/x0k/ps2-spy/internal/multi_loaders/ps2/world_population"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func startEventsClient(s *Setup, cfg *config.BotConfig, env string, settings ps2commands.SubscriptionSettings) *streaming.Client {
	client := streaming.NewClient(
		s.log,
		"wss://push.planetside2.com/streaming",
		env,
		cfg.ServiceId,
	)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		log := s.log.With(
			slog.String("component", "events_client_starter"),
			slog.String("platform", env),
		)
		retry.RetryWhileWithRecover(retry.Retryable{
			Try: func() error {
				err := client.Connect(s.ctx)
				if err != nil {
					log.Error("failed to connect to websocket", sl.Err(err))
					return err
				}
				defer func() {
					if err := client.Close(); err != nil {
						log.Error("failed to close websocket", sl.Err(err))
					}
				}()
				return client.Subscribe(s.ctx, settings)
			},
			While: retry.ContextIsNotCanceled,
			BeforeSleep: func(d time.Duration) {
				log.Debug("retry to connect", slog.Duration("after", d))
			},
		})
	}()
	return client
}

func mustSetupEventsPublisher(s *Setup, cfg *config.BotConfig) *ps2events.Publisher {
	eventsPublisher := ps2events.NewPublisher(s.log)
	msgPublisher := ps2messages.NewPublisher(s.log)
	serviceMsg := make(chan ps2messages.ServiceMessage[map[string]any])
	serviceMsgUnSub, err := msgPublisher.AddHandler(serviceMsg)
	if err != nil {
		s.log.Error("failed to add message handler", sl.Err(err))
		os.Exit(1)
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer serviceMsgUnSub()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-serviceMsg:
				eventsPublisher.Publish(msg.Payload)
			}
		}
	}()
	pc := startEventsClient(s, cfg, streaming.Ps2_env, ps2commands.SubscriptionSettings{
		Worlds: []string{"1", "10", "13", "17", "19", "40"},
		EventNames: []string{
			ps2events.PlayerLoginEventName,
			ps2events.PlayerLogoutEventName,
		},
	})
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case msg := <-pc.Msg:
				msgPublisher.Publish(msg)
			}
		}
	}()
	return eventsPublisher
}

type starter interface {
	Start()
	Stop()
}

func startInContext(s *Setup, starter starter) {
	starter.Start()
	s.wg.Add(1)
	context.AfterFunc(s.ctx, func() {
		defer s.wg.Done()
		starter.Stop()
	})
}

func startBot(s *Setup, cfg *config.BotConfig) {
	eventsPublisher := mustSetupEventsPublisher(s, cfg)
	httpClient := &http.Client{
		Timeout: cfg.HttpClientTimeout,
	}
	// loaders
	honuClient := honu.NewClient("https://wt.honu.pw", httpClient)
	startInContext(s, honuClient)
	fisuClient := fisu.NewClient("https://ps2.fisu.pw", httpClient)
	startInContext(s, fisuClient)
	voidWellClient := voidwell.NewClient("https://api.voidwell.com", httpClient)
	startInContext(s, voidWellClient)
	populationClient := population.NewClient("https://agg.ps2.live", httpClient)
	startInContext(s, populationClient)
	saerroClient := saerro.NewClient("https://saerro.ps2.live", httpClient)
	startInContext(s, saerroClient)
	ps2alertsClient := ps2alerts.NewClient("https://api.ps2alerts.com", httpClient)
	startInContext(s, ps2alertsClient)
	censusClient := census2.NewClient("https://census.daybreakgames.com", "", httpClient)
	sanctuaryClient := census2.NewClient("https://census.lithafalcon.cc", "", httpClient)
	// multi loaders
	popLoader := popMultiLoader.New(
		map[string]loaders.Loader[ps2.WorldsPopulation]{
			"honu":      worlds.NewHonuLoader(honuClient),
			"ps2live":   worlds.NewPS2LiveLoader(populationClient),
			"saerro":    worlds.NewSaerroLoader(saerroClient),
			"fisu":      worlds.NewFisuLoader(fisuClient),
			"sanctuary": worlds.NewSanctuaryLoader(sanctuaryClient),
			"voidwell":  worlds.NewVoidWellLoader(voidWellClient),
		},
		[]string{"honu", "ps2live", "saerro", "fisu", "sanctuary", "voidwell"},
	)
	startInContext(s, popLoader)
	worldPopLoader := worldPopMultiLoader.New(
		map[string]loaders.KeyedLoader[ps2.WorldId, ps2.DetailedWorldPopulation]{
			"honu":     world.NewHonuLoader(honuClient),
			"saerro":   world.NewSaerroLoader(saerroClient),
			"voidwell": world.NewVoidWellLoader(voidWellClient),
		},
		[]string{"honu", "saerro", "voidwell"},
	)
	startInContext(s, worldPopLoader)
	alertsLoader := alertsMultiLoader.New(
		map[string]loaders.Loader[ps2.Alerts]{
			"ps2alerts": alerts.NewPS2AlertsLoader(ps2alertsClient),
			"honu":      alerts.NewHonuLoader(honuClient),
			"census":    alerts.NewCensusLoader(censusClient),
			"voidwell":  alerts.NewVoidWellLoader(voidWellClient),
		},
		[]string{"ps2alerts", "honu", "census", "voidwell"},
	)
	startInContext(s, alertsLoader)
	worldAlertsLoader := alertsMultiLoader.NewWorldAlertsLoader(alertsLoader)
	startInContext(s, worldAlertsLoader)
	// bot
	botConfig := &bot.BotConfig{
		DiscordToken:          cfg.DiscordToken,
		CommandHandlerTimeout: cfg.CommandHandlerTimeout,
		Commands: bot.NewCommands(
			popLoader,
			worldPopLoader,
			alertsLoader,
		),
		Handlers: bot.NewHandlers(
			popLoader,
			worldPopLoader,
			alertsLoader,
			worldAlertsLoader,
		),
		EventsPublisher: eventsPublisher,
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
}
