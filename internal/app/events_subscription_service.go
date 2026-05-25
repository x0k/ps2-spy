package app

import (
	"context"
	"fmt"

	pubsub_adapters "github.com/x0k/ps2-spy/internal/adapters/pubsub"
	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/stats_tracker"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func newEventsSubscriptionService(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	ps module.Stopper,
	subs pubsub.SubscriptionsManager[events.EventType],
	charactersTracker *characters_tracker.Tracker,
	worldsTracker *worlds_tracker.WorldsTracker,
	statsTracker *stats_tracker.StatsTracker,
) module.Runnable {
	ch := pubsub_adapters.SubscribeToEvents(ps, subs)

	return module.NewRun(
		fmt.Sprintf("ps2.%s.events_subscription", platform),
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case e := <-ch.PlayerLogin:
					if err := charactersTracker.HandleLogin(ctx, platform, e); err != nil {
						log.Error(ctx, "failed to handle login event", sl.Err(err))
					}
				case e := <-ch.PlayerLogout:
					if err := charactersTracker.HandleLogout(ctx, platform, e); err != nil {
						log.Error(ctx, "failed to handle logout event", sl.Err(err))
					}
				case e := <-ch.AchievementEarned:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (achievementEarned)", sl.Err(err))
					}
				case e := <-ch.BattleRankUp:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (battleRankUp)", sl.Err(err))
					}
				case e := <-ch.Death:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err == nil {
						statsTracker.HandleDeathEvent(ctx, platform, e)
					} else {
						log.Error(ctx, "failed to handle world zone action (death)", sl.Err(err))
					}
				case e := <-ch.GainExperience:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err == nil {
						statsTracker.HandleGainExperienceEvent(ctx, platform, e)
					} else {
						log.Error(ctx, "failed to handle world zone action (gainExperience)", sl.Err(err))
					}
				case e := <-ch.ItemAdded:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (itemAdded)", sl.Err(err))
					}
				case e := <-ch.PlayerFacilityCapture:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (playerFacilityCapture)", sl.Err(err))
					}
				case e := <-ch.PlayerFacilityDefend:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (playerFacilityDefend)", sl.Err(err))
					}
				case e := <-ch.SkillAdded:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (skillAdded)", sl.Err(err))
					}
				case e := <-ch.VehicleDestroy:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (vehicleDestroy)", sl.Err(err))
					}

				case e := <-ch.MetagameEvent:
					if err := worldsTracker.HandleMetagameEvent(ctx, e); err != nil {
						log.Error(ctx, "failed to handle metagame event", sl.Err(err))
					}
				case e := <-ch.FacilityControl:
					if err := worldsTracker.HandleFacilityControl(ctx, e); err != nil {
						log.Error(ctx, "failed to handle facility control event", sl.Err(err))
					}
				case e := <-ch.ContinentLock:
					if err := worldsTracker.HandleContinentLock(ctx, e); err != nil {
						log.Error(ctx, "failed to handle continent lock event", sl.Err(err))
					}
				}
			}
		},
	)
}
