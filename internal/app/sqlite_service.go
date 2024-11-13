package app

import (
	"context"
	"database/sql"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sql_storage"

	_ "modernc.org/sqlite"
)

func newSqliteStorageService(log *logger.Logger, storagePath string, publisher pubsub.Publisher[storage.Event]) module.Service {
	return module.NewService("storage.sqlite", func(ctx context.Context) error {
		db, err := sql.Open("sqlite", storagePath)
		if err != nil {
			return err
		}
		storage, err := sql_storage.New(ctx, log, db, publisher)
		if err != nil {
			return err
		}
		<-ctx.Done()
		return storage.Close()
	})
}
