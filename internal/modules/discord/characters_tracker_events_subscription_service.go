package discord_module

import (
	"context"
	"fmt"
	"sync"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func newCharactersTrackerEventsSubscriptionService(
	ps module.PreStopper,
	platform ps2_platforms.Platform,
	charactersTrackerSubs pubsub.SubscriptionsManager[characters_tracker.EventType],
	eventsHandler *EventsHandler,
) module.Service {
	playerLogin := characters_tracker.Subscribe[characters_tracker.PlayerLogin](ps, charactersTrackerSubs)
	return module.NewService(
		fmt.Sprintf("discord.%s.characters_tracker_events_subscription", platform),
		func(ctx context.Context) error {
			wg := sync.WaitGroup{}
			for {
				select {
				case <-ctx.Done():
					wg.Wait()
					return nil
				case e := <-playerLogin:
					wg.Add(1)
					go func() {
						defer wg.Done()
						eventsHandler.HandlePlayerLogin(ctx, e)
					}()
				}
			}
		},
	)
}
