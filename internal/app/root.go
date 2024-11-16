package app

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/metrics"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
	"github.com/x0k/ps2-spy/internal/modules/discord/commands"
	ps2_module "github.com/x0k/ps2-spy/internal/modules/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
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

	discordModule, err := discord_module.New(log, &discord_module.Config{
		Token:                 cfg.Discord.Token,
		RemoveCommands:        cfg.Discord.RemoveCommands,
		CommandHandlerTimeout: cfg.Discord.CommandHandlerTimeout,
		EventHandlerTimeout:   cfg.Discord.EventHandlerTimeout,
		Commands:              commands.New(),
	})
	if err != nil {
		return nil, err
	}
	m.Append(discordModule)

	storagePubSub := pubsub.New[storage.EventType]()

	db, err := sql.Open("sqlite", cfg.Storage.Path)
	if err != nil {
		return nil, err
	}
	storage := sql_storage.New(log, db, storagePubSub)
	m.Append(module.NewService("storage.sqlite", storage.Start))

	httpClient := &http.Client{
		Timeout:   cfg.HttpClient.Timeout,
		Transport: metrics.InstrumentTransport(mt, metrics.DefaultTransportName, http.DefaultTransport),
	}

	ps2Module, err := ps2_module.New(&ps2_module.Config{
		Log:               log.With(slog.String("module", "ps2")),
		Metrics:           mt,
		Storage:           storage,
		AppName:           cfg.Ps2.AppName,
		Platform:          ps2_platforms.PC,
		StreamingEndpoint: cfg.Ps2.StreamingEndpoint,
		CensusServiceId:   cfg.Ps2.CensusServiceId,
		HttpClient:        httpClient,
	})
	if err != nil {
		return nil, err
	}
	m.Append(ps2Module)

	_ = ps2Module

	return m, nil
}
