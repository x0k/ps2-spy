package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

func startMetrics(ctx context.Context, log *logger.Logger, cfg config.MetricsConfig) {
	if !cfg.Enabled {
		log.Info(ctx, "metrics disabled")
		return
	}
	log.Info(ctx, "starting metrics", slog.String("address", cfg.Address))
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Error(ctx, "failed to start metrics", sl.Err(err))
		}
	}()
}
