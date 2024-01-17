package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

func startStorage(
	s *setup,
	cfg config.StorageConfig,
	publisher *storage.Publisher,
) (*sqlite.Storage, error) {
	const op = "startStorage"
	storage, err := sqlite.New(s.ctx, s.log, cfg.Path, publisher)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	err = storage.Start(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	s.wg.Add(1)
	context.AfterFunc(s.ctx, func() {
		defer s.wg.Done()
		if err := storage.Close(); err != nil {
			s.log.Error("cannot close storage", sl.Err(err))
		}
	})
	return storage, nil
}
