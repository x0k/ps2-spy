package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
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
	"github.com/x0k/ps2-spy/internal/loaders/ps2/alerts"
	"github.com/x0k/ps2-spy/internal/loaders/ps2/world"
	"github.com/x0k/ps2-spy/internal/loaders/ps2/worlds"
	alertsMultiLoader "github.com/x0k/ps2-spy/internal/multi_loaders/ps2/alerts"
	popMultiLoader "github.com/x0k/ps2-spy/internal/multi_loaders/ps2/population"
	worldPopMultiLoader "github.com/x0k/ps2-spy/internal/multi_loaders/ps2/world_population"
	"github.com/x0k/ps2-spy/internal/ps2"
)

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
