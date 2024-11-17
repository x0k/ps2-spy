package discord_module

import (
	"context"
	"fmt"
	"sync"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func publishEvent[E any](ctx context.Context, wg *sync.WaitGroup, log *logger.Logger, h func(context.Context, E) error, e E) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h(ctx, e); err != nil {
			log.Error(ctx, "failed to publish event", sl.Err(err))
		}
	}()
}

func newEventsSubscriptionService(
	log *logger.Logger,
	ps module.PreStopper,
	platform ps2_platforms.Platform,
	charactersTrackerSubs pubsub.SubscriptionsManager[characters_tracker.EventType],
	eventsHandler *Publisher,
) module.Service {
	playerLogin := characters_tracker.Subscribe[characters_tracker.PlayerLogin](ps, charactersTrackerSubs)
	return module.NewService(
		fmt.Sprintf("discord.%s.characters_tracker_events_subscription", platform),
		func(ctx context.Context) error {
			wg := &sync.WaitGroup{}
			for {
				select {
				case <-ctx.Done():
					wg.Wait()
					return nil
				case e := <-playerLogin:
					publishEvent(ctx, wg, log, eventsHandler.PublishPlayerLogin, e)
				}
			}
		},
	)
}
