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

	storagePubSub := pubsub.New[storage.EventType]()

	db, err := sql.Open("sqlite", cfg.Storage.Path)
	if err != nil {
		return nil, err
	}
	storage := sql_storage.New(
		log.With(slog.String("module", "storage")),
		db,
		storagePubSub,
	)
	m.Append(module.NewService("storage.sqlite", storage.Start))

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
		Transport: metrics.InstrumentTransport(
			mt,
			metrics.DefaultTransportName,
			http.DefaultTransport,
		),
	}

	ps2Module, err := ps2_module.New(
		log.With(slog.String("module", "ps2")),
		mt,
		storage,
		storagePubSub,
		httpClient,
		cfg.Ps2.StreamingEndpoint,
		cfg.Ps2.CensusServiceId,
		cfg.Ps2.AppName,
	)
	if err != nil {
		return nil, err
	}
	m.Append(ps2Module)

	discordModule, err := discord_module.New(
		log.With(slog.String("module", "discord")),
		cfg.Discord.Token,
		commands.New(),
		cfg.Discord.CommandHandlerTimeout,
		cfg.Discord.EventHandlerTimeout,
		cfg.Discord.RemoveCommands,
	)
	if err != nil {
		return nil, err
	}
	m.Append(discordModule)

	return m, nil
}
