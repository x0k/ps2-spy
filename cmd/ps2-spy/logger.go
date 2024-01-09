package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/x0k/ps2-spy/internal/config"
)

func mustSetupLogger(cfg *config.LoggerConfig) *slog.Logger {
	var level slog.Leveler
	switch cfg.Level {
	case config.DebugLevel:
		level = slog.LevelDebug
	case config.InfoLevel:
		level = slog.LevelInfo
	case config.WarnLevel:
		level = slog.LevelWarn
	case config.ErrorLevel:
		level = slog.LevelError
	default:
		log.Fatalf("Unknown level: %s, expect %q, %q, %q or %q", cfg.Level, config.DebugLevel, config.InfoLevel, config.WarnLevel, config.ErrorLevel)
	}
	options := &slog.HandlerOptions{
		Level: level,
	}
	var handler slog.Handler
	switch cfg.HandlerType {
	case config.TextHandler:
		handler = slog.NewTextHandler(os.Stdout, options)
	case config.JSONHandler:
		handler = slog.NewJSONHandler(os.Stdout, options)
	default:
		log.Fatalf("Unknown handler type: %s, expect %q or %q", cfg.HandlerType, config.TextHandler, config.JSONHandler)
	}
	return slog.New(handler)
}
