package app

import (
	"log"
	"os"

	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/handlers/slogpretty"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

const (
	TextHandler   = "text"
	JSONHandler   = "json"
	PrettyHandler = "pretty"
)

type LoggerConfig struct {
	Level       string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
	HandlerType string `yaml:"handler_type" env:"LOGGER_HANDLER_TYPE" env-default:"text"`
}

func MustNewLogger(cfg *LoggerConfig) *logger.Logger {
	var level slog.Leveler
	switch cfg.Level {
	case DebugLevel:
		level = slog.LevelDebug
	case InfoLevel:
		level = slog.LevelInfo
	case WarnLevel:
		level = slog.LevelWarn
	case ErrorLevel:
		level = slog.LevelError
	default:
		log.Fatalf("Unknown level: %s, expect %q, %q, %q or %q", cfg.Level, DebugLevel, InfoLevel, WarnLevel, ErrorLevel)
	}
	options := &slog.HandlerOptions{
		Level: level,
	}
	var handler slog.Handler
	switch cfg.HandlerType {
	case TextHandler:
		handler = slog.NewTextHandler(os.Stdout, options)
	case JSONHandler:
		handler = slog.NewJSONHandler(os.Stdout, options)
	case PrettyHandler:
		otps := &slogpretty.PrettyHandlerOptions{
			SlogOpts: options,
		}
		handler = otps.NewPrettyHandler(os.Stdout)
	default:
		log.Fatalf("Unknown handler type: %s, expect %q or %q", cfg.HandlerType, TextHandler, JSONHandler)
	}
	return logger.New(slog.New(handler))
}
