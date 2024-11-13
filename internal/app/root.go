package app

import (
	"net/http"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	discord_module "github.com/x0k/ps2-spy/internal/modules/discord"
	"github.com/x0k/ps2-spy/internal/modules/discord/commands"
	profiler_module "github.com/x0k/ps2-spy/internal/modules/profiler"
	ps2_module "github.com/x0k/ps2-spy/internal/modules/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

func NewRoot(cfg *Config, log *logger.Logger) (*module.Root, error) {
	m := module.NewRoot(log.Logger)

	profilerModule := profiler_module.New(&profiler_module.Config{
		Enabled: cfg.Profiler.Enabled,
		Address: cfg.Profiler.Address,
	}, log)
	m.Append(profilerModule)

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
	storageService := newSqliteStorageService(log, cfg.Storage.Path, storagePubSub)
	m.Append(storageService)

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
	}
	_ = httpClient

	ps2Module, err := ps2_module.New(log, &ps2_module.Config{
		Platform:          platforms.PC,
		StreamingEndpoint: cfg.Ps2.StreamingEndpoint,
		CensusServiceId:   cfg.Ps2.CensusServiceId,
	})
	if err != nil {
		return nil, err
	}
	m.Append(ps2Module)

	_ = ps2Module

	return m, nil
}
