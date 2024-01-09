package main

import (
	"context"
	"os"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

func mustSetupStorage(s *Setup, cfg *config.StorageConfig) *sqlite.Storage {
	storage, err := sqlite.New(s.ctx, s.log, cfg.Path)
	if err != nil {
		s.log.Error("cannot open storage", sl.Err(err))
		os.Exit(1)
	}
	err = storage.Start(s.ctx)
	if err != nil {
		s.log.Error("cannot start storage", sl.Err(err))
		os.Exit(1)
	}
	s.wg.Add(1)
	context.AfterFunc(s.ctx, func() {
		defer s.wg.Done()
		if err := storage.Close(); err != nil {
			s.log.Error("cannot close storage", sl.Err(err))
		}
	})
	return storage
}
