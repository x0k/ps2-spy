package app

import (
	"net/http"
	"net/http/pprof"

	http_adapters "github.com/x0k/ps2-spy/internal/adapters/http"
	"github.com/x0k/ps2-spy/internal/lib/module"
)

func newProfilerService(
	address string,
	fataler module.Fataler,
) module.Runnable {
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	return http_adapters.NewService("profiler", srv, fataler)
}
