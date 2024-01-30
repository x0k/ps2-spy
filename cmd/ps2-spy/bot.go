package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/meta"
)

func startNewBot(
	ctx context.Context,
	cfg *bot.BotConfig,
	pcEventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	pcEventHandlers bot.EventHandlers,
	ps4euEventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	ps4euEventHandlers bot.EventHandlers,
	ps4usEventTrackingChannelsLoader loaders.QueriedLoader[any, []meta.ChannelId],
	ps4usEventHandlers bot.EventHandlers,
) error {
	const op = "startBot"
	log := infra.OpLogger(ctx, op)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		retryable.New(
			func(ctx context.Context) error {
				b, err := bot.New(ctx, cfg)
				if err != nil {
					return err
				}
				defer func() {
					if err := b.Stop(ctx); err != nil {
						log.Error("failed to stop bot", sl.Err(err))
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
			perform.Log(log, slog.LevelError, "bot failed, restarting"),
		)
	}()
	return nil
}
