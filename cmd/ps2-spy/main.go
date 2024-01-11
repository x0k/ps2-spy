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
)

var (
	config_path string
)

func init() {
	flag.StringVar(&config_path, "config", os.Getenv("CONFIG_PATH"), "Config path")
	flag.Parse()
}

type Setup struct {
	log *slog.Logger
	ctx context.Context
	wg  *sync.WaitGroup
}

func main() {
	cfg := config.MustLoad(config_path)
	log := mustSetupLogger(&cfg.Logger)
	log.Info("starting...", slog.String("log_level", cfg.Logger.Level))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := &Setup{
		log: log,
		ctx: ctx,
		wg:  &sync.WaitGroup{},
	}

	storage := mustSetupStorage(s, &cfg.Storage)
	_ = storage

	startBot(s, &cfg.Bot)
	startPrinter(s, &cfg.Bot)

	log.Info("Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("gracefully shutting down.")
	cancel()
	s.wg.Wait()
}
