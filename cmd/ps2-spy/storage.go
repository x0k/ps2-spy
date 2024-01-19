package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

func startStorage(
	ctx context.Context,
	cfg config.StorageConfig,
	publisher *storage.Publisher,
) (*sqlite.Storage, error) {
	const op = "startStorage"
	storage, err := sqlite.New(ctx, cfg.Path, publisher)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = storage.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	wg := infra.Wg(ctx)
	wg.Add(1)
	context.AfterFunc(ctx, func() {
		defer wg.Done()
		if err := storage.Close(ctx); err != nil {
			infra.OpLogger(ctx, op).Error("cannot close storage", sl.Err(err))
		}
	})
	return storage, nil
}
