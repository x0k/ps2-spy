package discord_module

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func newEventsSubscriptionService(
	ps module.PostStopper,
	platform ps2_platforms.Platform,
	charactersTrackerSubs pubsub.SubscriptionsManager[characters_tracker.EventType],
	storageSubs pubsub.SubscriptionsManager[storage.EventType],
	worldsTrackerSubs pubsub.SubscriptionsManager[worlds_tracker.EventType],
	handlersManager *discord_events.HandlersManager,
) module.Service {
	playerLogin := characters_tracker.Subscribe[characters_tracker.PlayerLogin](ps, charactersTrackerSubs)
	playerLogout := characters_tracker.Subscribe[characters_tracker.PlayerLogout](ps, charactersTrackerSubs)
	outfitMembersUpdate := storage.Subscribe[storage.OutfitMembersUpdate](ps, storageSubs)
	facilityControl := worlds_tracker.Subscribe[worlds_tracker.FacilityControl](ps, worldsTrackerSubs)
	facilityLoss := worlds_tracker.Subscribe[worlds_tracker.FacilityLoss](ps, worldsTrackerSubs)
	return module.NewService(
		fmt.Sprintf("discord.%s.events_subscription", platform),
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case e := <-playerLogin:
					handlersManager.HandlePlayerLogin(ctx, e)
				case e := <-playerLogout:
					handlersManager.HandlePlayerLogout(ctx, e)
				case e := <-outfitMembersUpdate:
					if e.Platform == platform {
						handlersManager.HandleOutfitMembersUpdate(ctx, e)
					}
				case e := <-facilityControl:
					handlersManager.HandleFacilityControl(ctx, e)
				case e := <-facilityLoss:
					handlersManager.HandleFacilityLoss(ctx, e)
				}
			}
		},
	)
}
