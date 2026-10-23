package app

import (
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/migrator"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/storage"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

func newStorage(log *logger.Logger, cfg *Config, m *module.Module, storePubSub pubsub.Publisher[storage.Event]) *sql_storage.Storage {
	mig := migrator.New(
		log.Logger.With(sl.Component("migrator")),
		cfg.Storage.Path,
		cfg.Storage.MigrationsPath,
	)
	m.OnStart(module.NewRun("migrator", mig.Migrate))

	store := sql_storage.New(
		log.With(sl.Component("storage")),
		cfg.Storage.Path,
		storePubSub,
	)
	m.OnStart(module.NewRun("storage", store.Open))
	m.OnStop(module.NewRun("storage", store.Close))

	return store
}
