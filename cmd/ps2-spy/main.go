package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
)

var (
	config_path string
)

func init() {
	flag.StringVar(&config_path, "config", os.Getenv("CONFIG_PATH"), "Config path")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	cfg := config.MustLoad(config_path)

	log := mustSetupLogger(&cfg.Logger)
	log.Info("starting...", slog.String("log_level", cfg.Logger.Level))
	ctx = context.WithValue(ctx, infra.LoggerKey, log)

	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, infra.WgKey, wg)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	start(ctx, cfg)

	log.Info("Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("gracefully shutting down.")
	cancel()
	wg.Wait()
}
