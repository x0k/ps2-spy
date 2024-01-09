package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
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
	cfg := config.MustLoad(config_path)
	log := mustSetupLogger(&cfg.LoggerConfig)

	log.Info("starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storage := mustSetupStorage(ctx, &cfg.Storage, log)
	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("cannot close storage", sl.Err(err))
		}
	}()
	err := storage.Migrate(ctx)
	if err != nil {
		log.Error("cannot migrate storage", sl.Err(err))
		os.Exit(1)
	}

	ps2Service := setupPs2Service(ctx)
	ps2Service.Start()
	defer ps2Service.Stop()

	ps2events := streaming.NewClient("wss://push.planetside2.com/streaming", streaming.Ps2_env, "example")
	err = ps2events.Connect(ctx)
	if err != nil {
		log.Error("failed to connect to websocket", sl.Err(err))
	} else {
		defer ps2events.Close()
	}

	b := mustSetupBot(ctx, &cfg.Bot, log, ps2Service)
	defer b.Stop()

	log.Info("bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("gracefully shutting down.")
}
