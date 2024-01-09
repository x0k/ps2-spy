package main

import (
	"context"
	"os"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/config"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func mustSetupBot(s *Setup, cfg *config.BotConfig, service *ps2.Service) *bot.Bot {
	b, err := bot.New(s.ctx, &bot.BotConfig{
		DiscordToken:          cfg.DiscordToken,
		CommandHandlerTimeout: cfg.CommandHandlerTimeout,
	}, s.log, service)
	if err != nil {
		s.log.Error("failed to create bot", sl.Err(err))
		os.Exit(1)
	}
	s.wg.Add(1)
	context.AfterFunc(s.ctx, func() {
		defer s.wg.Done()
		if err := b.Stop(); err != nil {
			s.log.Error("failed to stop bot", sl.Err(err))
		}
	})
	return b
}
