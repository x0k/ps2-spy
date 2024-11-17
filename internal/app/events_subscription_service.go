package app

import (
	"context"
	"fmt"
	"sync"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func newEventsSubscriptionService(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	ps module.PreStopper,
	subs pubsub.SubscriptionsManager[events.EventType],
	charactersTracker *characters_tracker.CharactersTracker,
	worldsTracker *worlds_tracker.WorldsTracker,
) module.Service {
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

	metagameEvent := census2_adapters.Subscribe[events.MetagameEvent](ps, subs)
	facilityControl := census2_adapters.Subscribe[events.FacilityControl](ps, subs)
	continentLock := census2_adapters.Subscribe[events.ContinentLock](ps, subs)

	return module.NewService(
		fmt.Sprintf("ps2.%s.events_subscription", platform),
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

				case e := <-metagameEvent:
					if err := worldsTracker.HandleMetagameEvent(ctx, e); err != nil {
						log.Error(ctx, "failed to handle metagame event", sl.Err(err))
					}
				case e := <-facilityControl:
					if err := worldsTracker.HandleFacilityControl(ctx, e); err != nil {
						log.Error(ctx, "failed to handle facility control event", sl.Err(err))
					}
				case e := <-continentLock:
					if err := worldsTracker.HandleContinentLock(ctx, e); err != nil {
						log.Error(ctx, "failed to handle continent lock event", sl.Err(err))
					}
				}
			}
		},
	)
}
