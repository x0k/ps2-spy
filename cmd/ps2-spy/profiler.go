package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"net/http/pprof"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
)

func startProfiler(ctx context.Context, wg *sync.WaitGroup, log *logger.Logger, cfg config.ProfilerConfig) {
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
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: mux,
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(ctx, "failed to start profiler", sl.Err(err))
		}
	}()
	context.AfterFunc(ctx, func() {
		defer wg.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error(ctx, "failed to shutdown profiler", sl.Err(err))
		}
	})
}
