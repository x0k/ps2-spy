package app

import (
	"net/http"

	"github.com/x0k/ps2-spy/internal/discord_module"
	"github.com/x0k/ps2-spy/internal/discord_module/commands"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/profiler_module"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
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
	storageService := sqlite.NewService(log, cfg.Storage.Path, storagePubSub)
	m.Append(storageService)

	httpClient := &http.Client{
		Timeout: cfg.HttpClient.Timeout,
	}
	_ = httpClient

	return m, nil
}
