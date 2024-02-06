package main

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/metrics"
)

func startMetrics(ctx context.Context, wg *sync.WaitGroup, log *logger.Logger, cfg config.MetricsConfig) metrics.Metrics {
	if !cfg.Enabled {
		log.Info(ctx, "metrics disabled")
		return metrics.NewStub()
	}
	log.Info(ctx, "starting metrics", slog.String("address", cfg.Address))
	m := metrics.New("ps2spy")
	reg := prometheus.NewRegistry()
	mux := http.NewServeMux()
	m.Register(reg)
	reg.MustRegister(collectors.NewGoCollector())
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Error(ctx, "failed to start metrics", sl.Err(err))
		}
	}()
	return m
}
