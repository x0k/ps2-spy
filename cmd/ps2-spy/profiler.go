package main

import (
	"context"
	"log/slog"
	"net/http"

	"net/http/pprof"

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
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	go func() {
		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Error(ctx, "failed to start profiler", sl.Err(err))
		}
	}()
}
