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
	ps module.PostStopper,
	subs pubsub.SubscriptionsManager[events.EventType],
	charactersTracker *characters_tracker.Tracker,
	worldsTracker *worlds_tracker.WorldsTracker,
	statsTracker *stats_tracker.StatsTracker,
) module.Runnable {
	playerLogin := pubsub_adapters.SubscribeTo[events.EventType, events.PlayerLogin](ps, subs)
	playerLogout := pubsub_adapters.SubscribeTo[events.EventType, events.PlayerLogout](ps, subs)
	achievementEarned := pubsub_adapters.SubscribeTo[events.EventType, events.AchievementEarned](ps, subs)
	battleRankUp := pubsub_adapters.SubscribeTo[events.EventType, events.BattleRankUp](ps, subs)
	death := pubsub_adapters.SubscribeTo[events.EventType, events.Death](ps, subs)
	gainExperience := pubsub_adapters.SubscribeTo[events.EventType, events.GainExperience](ps, subs)
	itemAdded := pubsub_adapters.SubscribeTo[events.EventType, events.ItemAdded](ps, subs)
	playerFacilityCapture := pubsub_adapters.SubscribeTo[events.EventType, events.PlayerFacilityCapture](ps, subs)
	playerFacilityDefend := pubsub_adapters.SubscribeTo[events.EventType, events.PlayerFacilityDefend](ps, subs)
	skillAdded := pubsub_adapters.SubscribeTo[events.EventType, events.SkillAdded](ps, subs)
	vehicleDestroy := pubsub_adapters.SubscribeTo[events.EventType, events.VehicleDestroy](ps, subs)

	metagameEvent := pubsub_adapters.SubscribeTo[events.EventType, events.MetagameEvent](ps, subs)
	facilityControl := pubsub_adapters.SubscribeTo[events.EventType, events.FacilityControl](ps, subs)
	continentLock := pubsub_adapters.SubscribeTo[events.EventType, events.ContinentLock](ps, subs)

	return module.NewRun(
		fmt.Sprintf("ps2.%s.events_subscription", platform),
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case e := <-playerLogin:
					if err := charactersTracker.HandleLogin(ctx, platform, e); err != nil {
						log.Error(ctx, "failed to handle login event", sl.Err(err))
					}
				case e := <-playerLogout:
					if err := charactersTracker.HandleLogout(ctx, platform, e); err != nil {
						log.Error(ctx, "failed to handle logout event", sl.Err(err))
					}
				case e := <-achievementEarned:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (achievementEarned)", sl.Err(err))
					}
				case e := <-battleRankUp:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (battleRankUp)", sl.Err(err))
					}
				case e := <-death:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err == nil {
						statsTracker.HandleDeathEvent(ctx, platform, e)
					} else {
						log.Error(ctx, "failed to handle world zone action (death)", sl.Err(err))
					}
				case e := <-gainExperience:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err == nil {
						statsTracker.HandleGainExperienceEvent(ctx, platform, e)
					} else {
						log.Error(ctx, "failed to handle world zone action (gainExperience)", sl.Err(err))
					}
				case e := <-itemAdded:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (itemAdded)", sl.Err(err))
					}
				case e := <-playerFacilityCapture:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (playerFacilityCapture)", sl.Err(err))
					}
				case e := <-playerFacilityDefend:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (playerFacilityDefend)", sl.Err(err))
					}
				case e := <-skillAdded:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (skillAdded)", sl.Err(err))
					}
				case e := <-vehicleDestroy:
					if err := charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID); err != nil {
						log.Error(ctx, "failed to handle world zone action (vehicleDestroy)", sl.Err(err))
					}

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
