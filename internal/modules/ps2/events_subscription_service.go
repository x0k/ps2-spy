package ps2_module

import (
	"context"
	"fmt"
	"sync"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func newEventsSubscriptionService(
	platform ps2_platforms.Platform,
	ps module.PreStopper,
	subs pubsub.SubscriptionsManager[events.EventType],
	charactersTracker *characters_tracker.CharactersTracker,
) module.Service {
	return module.NewService(
		fmt.Sprintf("events_subscription.%s", platform),
		func(ctx context.Context) error {
			wg := sync.WaitGroup{}
			playerLogin := census2_adapters.Subscribe[events.PlayerLogin](ps, subs)
			playerLogout := census2_adapters.Subscribe[events.PlayerLogout](ps, subs)
			achievementEarned := census2_adapters.Subscribe[events.AchievementEarned](ps, subs)
			battleRankUp := census2_adapters.Subscribe[events.BattleRankUp](ps, subs)
			death := census2_adapters.Subscribe[events.Death](ps, subs)
			gainExperience := census2_adapters.Subscribe[events.GainExperience](ps, subs)
			itemAdded := census2_adapters.Subscribe[events.ItemAdded](ps, subs)
			playerFacilityCapture := census2_adapters.Subscribe[events.PlayerFacilityCapture](ps, subs)
			playerFacilityDefend := census2_adapters.Subscribe[events.PlayerFacilityDefend](ps, subs)
			skillAdded := census2_adapters.Subscribe[events.SkillAdded](ps, subs)
			vehicleDestroy := census2_adapters.Subscribe[events.VehicleDestroy](ps, subs)
			for {
				select {
				case <-ctx.Done():
					wg.Wait()
					return nil
				case e := <-playerLogin:
					wg.Add(1)
					go func() {
						defer wg.Done()
						charactersTracker.HandleLogin(ctx, e)
					}()
				case e := <-playerLogout:
					charactersTracker.HandleLogout(ctx, e)
				case e := <-achievementEarned:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-battleRankUp:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-death:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-gainExperience:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-itemAdded:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-playerFacilityCapture:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-playerFacilityDefend:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-skillAdded:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-vehicleDestroy:
					charactersTracker.HandleWorldZoneAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
				}
			}
		},
	)
}
