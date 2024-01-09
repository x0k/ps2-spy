package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func mustSetupBot(ctx context.Context, cfg *config.BotConfig, log *slog.Logger, service *ps2.Service) *bot.Bot {
	b, err := bot.New(ctx, &bot.BotConfig{
		DiscordToken:          cfg.DiscordToken,
		CommandHandlerTimeout: cfg.CommandHandlerTimeout,
	}, log, service)
	if err != nil {
		log.Error("failed to create bot", sl.Err(err))
		os.Exit(1)
	}
	return b
}
