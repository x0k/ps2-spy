package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	http_adapters "github.com/x0k/ps2-spy/internal/adapters/http"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/metrics"
)

func newMetricsService(
	m *metrics.Metrics,
	address string,
	fataler module.Fataler,
) module.Runnable {
	mux := http.NewServeMux()
	reg := prometheus.NewRegistry()
	m.Register(reg)
	reg.MustRegister(collectors.NewGoCollector())
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	return http_adapters.NewService("metrics", srv, fataler)
}
