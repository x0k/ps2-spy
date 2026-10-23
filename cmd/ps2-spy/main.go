package main

import (
	"context"
	"flag"
	"os"

	"log/slog"

	"github.com/x0k/ps2-spy/internal/app"
	"github.com/x0k/ps2-spy/internal/lib/logger"
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
	if err := run(ctx, cfg, log); err != nil {
		log.Error(ctx, "fatal error", sl.Err(err))
		log.Info(ctx, "application stopped")
		os.Exit(1)
	}
	log.Info(ctx, "application stopped")
}

func run(ctx context.Context, cfg *app.Config, log *logger.Logger) error {
	root, err := app.NewRoot(cfg, log)
	if err != nil {
		return err
	}
	return root.Run(ctx)
}
