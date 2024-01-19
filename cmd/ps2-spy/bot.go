package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
)

func startBot(
	ctx context.Context,
	cfg *bot.BotConfig,
) error {
	const op = "startBot"
	log := infra.OpLogger(ctx, op)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		retry.RetryWhileWithRecover(retry.Retryable{
			Try: func() error {
				b, err := bot.New(ctx, cfg)
				if err != nil {
					return err
				}
				defer func() {
					if err := b.Stop(ctx); err != nil {
						log.Error("failed to stop bot", sl.Err(err))
					}
				}()
				<-ctx.Done()
				return ctx.Err()
			},
			While: retry.ContextIsNotCanceled,
			BeforeSleep: func(d time.Duration) {
				log.Debug("retry to start bot", slog.Duration("after", d))
			},
		})
	}()
	return nil
}
