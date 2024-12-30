package app

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	census_data_provider "github.com/x0k/ps2-spy/internal/data_providers/census"
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
	playerLogin := census_data_provider.Subscribe[events.PlayerLogin](ps, subs)
	playerLogout := census_data_provider.Subscribe[events.PlayerLogout](ps, subs)
	achievementEarned := census_data_provider.Subscribe[events.AchievementEarned](ps, subs)
	battleRankUp := census_data_provider.Subscribe[events.BattleRankUp](ps, subs)
	death := census_data_provider.Subscribe[events.Death](ps, subs)
	gainExperience := census_data_provider.Subscribe[events.GainExperience](ps, subs)
	itemAdded := census_data_provider.Subscribe[events.ItemAdded](ps, subs)
	playerFacilityCapture := census_data_provider.Subscribe[events.PlayerFacilityCapture](ps, subs)
	playerFacilityDefend := census_data_provider.Subscribe[events.PlayerFacilityDefend](ps, subs)
	skillAdded := census_data_provider.Subscribe[events.SkillAdded](ps, subs)
	vehicleDestroy := census_data_provider.Subscribe[events.VehicleDestroy](ps, subs)

	metagameEvent := census_data_provider.Subscribe[events.MetagameEvent](ps, subs)
	facilityControl := census_data_provider.Subscribe[events.FacilityControl](ps, subs)
	continentLock := census_data_provider.Subscribe[events.ContinentLock](ps, subs)

	return module.NewRun(
		fmt.Sprintf("ps2.%s.events_subscription", platform),
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case e := <-playerLogin:
					charactersTracker.HandleLogin(ctx, platform, e)
				case e := <-playerLogout:
					charactersTracker.HandleLogout(ctx, platform, e)
				case e := <-achievementEarned:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-battleRankUp:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-death:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
					statsTracker.HandleDeathEvent(ctx, platform, e)
				case e := <-gainExperience:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
					statsTracker.HandleGainExperienceEvent(ctx, platform, e)
				case e := <-itemAdded:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-playerFacilityCapture:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-playerFacilityDefend:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-skillAdded:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)
				case e := <-vehicleDestroy:
					charactersTracker.HandleWorldZoneAction(ctx, platform, e.WorldID, e.ZoneID, e.CharacterID)

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
