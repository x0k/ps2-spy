package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

func startProfiler(ctx context.Context, log *logger.Logger, cfg config.ProfilerConfig) {
	if !cfg.Enabled {
		log.Info(ctx, "profiler disabled")
		return
	}
	log.Info(ctx, "starting profiler", slog.String("address", cfg.Address))
	go func() {
		if err := http.ListenAndServe(cfg.Address, nil); err != nil {
			log.Error(ctx, "failed to start profiler", sl.Err(err))
		}
	}()
}
