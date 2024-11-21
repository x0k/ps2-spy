package main

import (
	"context"
	"flag"
	"os"

	"log/slog"

	"github.com/x0k/ps2-spy/internal/app"
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
	cfg := app.MustLoadConfig(config_path)
	log := app.MustNewLogger(&cfg.Logger)
	ctx := context.Background()
	log.Info(ctx, "starting application", slog.String("log_level", cfg.Logger.Level))
	root, err := app.NewRoot(cfg, log)
	if err != nil {
		log.Error(ctx, "failed to run", sl.Err(err))
		return
	}
	if err := root.Start(ctx); err != nil {
		log.Error(ctx, "fatal error", sl.Err(err))
	}
	log.Info(ctx, "application stopped")
}
