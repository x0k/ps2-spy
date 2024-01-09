package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

func mustSetupStorage(ctx context.Context, cfg *config.StorageConfig, log *slog.Logger) *sqlite.Storage {
	storage, err := sqlite.New(ctx, log, cfg.Path)
	if err != nil {
		log.Error("cannot open storage", sl.Err(err))
		os.Exit(1)
	}
	return storage
}
