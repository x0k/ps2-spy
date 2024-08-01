package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/bot"
	"github.com/x0k/ps2-spy/internal/cache/facility_cache"
	"github.com/x0k/ps2-spy/internal/cache/outfit_cache"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/loaders/facility_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/metrics"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
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

func newEventHandler(
	log *logger.Logger,
	_ metrics.Metrics,
	platform platforms.Platform,
	charactersTrackerPublisher *characters_tracker.Publisher,
	outfitMembersSaverPublisher *outfit_members_saver.Publisher,
	worldsTrackerPublisher *worlds_tracker.Publisher,
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	facilityCache *facility_cache.StorageCache,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	charactersLoader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
) bot.EventHandlers {
	pLog := log.With(slog.String("platform", string(platform)))
	outfitLoader := loaders.NewCachedQueriedLoader(
		pLog.Logger,
		outfit_loader.NewCensus(censusClient, platform),
		outfit_cache.NewStorage(pLog, sqlStorage, platform),
	)
	facilityLoader := loaders.NewCachedQueriedLoader(
		pLog.Logger,
		facility_loader.NewCensus(censusClient, platforms.PlatformNamespace(platform)),
		facilityCache,
	)
	return bot.NewEventHandlers(
		pLog,
		charactersTrackerPublisher,
		outfitMembersSaverPublisher,
		worldsTrackerPublisher,
		characterLoader,
		outfitLoader,
		facilityLoader,
		charactersLoader,
	)
}
