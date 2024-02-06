package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/meta"
)

func startNewBot(
	ctx context.Context,
	log *logger.Logger,
	cfg *bot.BotConfig,
	pcEventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	pcEventHandlers bot.EventHandlers,
	ps4euEventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	ps4euEventHandlers bot.EventHandlers,
	ps4usEventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	ps4usEventHandlers bot.EventHandlers,
) {
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = retryable.New(
			func(ctx context.Context) error {
				b, err := bot.New(ctx, log, cfg)
				if err != nil {
					return err
				}
				defer func() {
					if err := b.Stop(ctx); err != nil {
						log.Error(ctx, "failed to stop bot", sl.Err(err))
					}
				}()
				if err := b.StartEventHandlers(ctx, pcEventTrackingChannelsLoader, pcEventHandlers); err != nil {
					return err
				}
				if err := b.StartEventHandlers(ctx, ps4euEventTrackingChannelsLoader, ps4euEventHandlers); err != nil {
					return err
				}
				if err := b.StartEventHandlers(ctx, ps4usEventTrackingChannelsLoader, ps4usEventHandlers); err != nil {
					return err
				}
				<-ctx.Done()
				return ctx.Err()
			},
		).Run(
			ctx,
			while.ContextIsNotCancelled,
			perform.RecoverSuspenseDuration(1*time.Second),
			perform.Log(log.Logger, slog.LevelError, "bot failed, restarting"),
		)
	}()
}
