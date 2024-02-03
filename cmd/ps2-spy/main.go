package main

import (
	"context"
	"flag"
	"log/slog"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
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
	log.Info(ctx, "starting...", slog.String("log_level", cfg.Logger.Level))
	ctx = context.WithValue(ctx, infra.LoggerKey, log)

	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, infra.WgKey, wg)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := start(ctx, log, cfg)
	if err != nil {
		log.Error(ctx, "unsuccess start, shutting down", sl.Err(err))
		cancel()
		wg.Wait()
		os.Exit(1)
	}

	log.Info(ctx, "press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info(ctx, "gracefully shutting down.")
	cancel()
	wg.Wait()
	log.Info(ctx, "shutdown complete.")
}
