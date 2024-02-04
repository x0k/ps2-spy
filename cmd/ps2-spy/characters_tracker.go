package main

import (
	"context"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func startCharactersTracker(
	ctx context.Context,
	charactersTracker *characters_tracker.CharactersTracker,
	ps2EventsPublisher *ps2events.Publisher,
) error {
	const op = "startCharactersTracker"
	charactersTracker.Start(ctx)
	achievementEarned := make(chan ps2events.AchievementEarned)
	achievementUnSub := ps2EventsPublisher.AddAchievementEarnedHandler(achievementEarned)
	battleRankUp := make(chan ps2events.BattleRankUp)
	battleRankUpUnSub := ps2EventsPublisher.AddBattleRankUpHandler(battleRankUp)
	death := make(chan ps2events.Death)
	deathUnSub := ps2EventsPublisher.AddDeathHandler(death)
	gainExperience := make(chan ps2events.GainExperience)
	gainExperienceUnSub := ps2EventsPublisher.AddGainExperienceHandler(gainExperience)
	itemAdded := make(chan ps2events.ItemAdded)
	itemAddedUnSub := ps2EventsPublisher.AddItemAddedHandler(itemAdded)
	playerFacilityCapture := make(chan ps2events.PlayerFacilityCapture)
	playerFacilityCaptureUnSub := ps2EventsPublisher.AddPlayerFacilityCaptureHandler(playerFacilityCapture)
	playerFacilityDefend := make(chan ps2events.PlayerFacilityDefend)
	playerFacilityDefendUnSub := ps2EventsPublisher.AddPlayerFacilityDefendHandler(playerFacilityDefend)
	playerLogin := make(chan ps2events.PlayerLogin)
	playerLoginUnSub := ps2EventsPublisher.AddPlayerLoginHandler(playerLogin)
	playerLogout := make(chan ps2events.PlayerLogout)
	playerLogoutUnSub := ps2EventsPublisher.AddPlayerLogoutHandler(playerLogout)
	skillAdded := make(chan ps2events.SkillAdded)
	skillAddedUnSub := ps2EventsPublisher.AddSkillAddedHandler(skillAdded)
	vehicleDestroy := make(chan ps2events.VehicleDestroy)
	vehicleDestroyUnSub := ps2EventsPublisher.AddVehicleDestroyHandler(vehicleDestroy)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			achievementUnSub()
			battleRankUpUnSub()
			deathUnSub()
			gainExperienceUnSub()
			itemAddedUnSub()
			playerFacilityCaptureUnSub()
			playerFacilityDefendUnSub()
			playerLoginUnSub()
			playerLogoutUnSub()
			skillAddedUnSub()
			vehicleDestroyUnSub()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-playerLogin:
				wg.Add(1)
				go charactersTracker.HandleLoginTask(ctx, wg, e)
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
	}()
	return nil
}

func startNewCharactersTracker(
	ctx context.Context,
	log *logger.Logger,
	worldIds []ps2.WorldId,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	ps2EventsPublisher *ps2events.Publisher,
	charactersTrackerPublisher *characters_tracker.Publisher,
) (*characters_tracker.CharactersTracker, error) {
	charactersTracker := characters_tracker.New(log, worldIds, characterLoader, charactersTrackerPublisher)
	return charactersTracker, startCharactersTracker(
		ctx,
		charactersTracker,
		ps2EventsPublisher,
	)
}
