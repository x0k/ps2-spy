package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func startCharactersTracker(
	ctx context.Context,
	charactersTracker *characters_tracker.CharactersTracker,
	ps2EventsPublisher *publisher.Publisher,
) error {
	const op = "startCharactersTracker"
	charactersTracker.Start(ctx)
	achievementEarned := make(chan ps2events.AchievementEarned)
	achievementUnSub, err := ps2EventsPublisher.AddHandler(achievementEarned)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	battleRankUp := make(chan ps2events.BattleRankUp)
	battleRankUpUnSub, err := ps2EventsPublisher.AddHandler(battleRankUp)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	death := make(chan ps2events.Death)
	deathUnSub, err := ps2EventsPublisher.AddHandler(death)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	gainExperience := make(chan ps2events.GainExperience)
	gainExperienceUnSub, err := ps2EventsPublisher.AddHandler(gainExperience)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	itemAdded := make(chan ps2events.ItemAdded)
	itemAddedUnSub, err := ps2EventsPublisher.AddHandler(itemAdded)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	playerFacilityCapture := make(chan ps2events.PlayerFacilityCapture)
	playerFacilityCaptureUnSub, err := ps2EventsPublisher.AddHandler(playerFacilityCapture)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	playerFacilityDefend := make(chan ps2events.PlayerFacilityDefend)
	playerFacilityDefendUnSub, err := ps2EventsPublisher.AddHandler(playerFacilityDefend)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	playerLogin := make(chan ps2events.PlayerLogin)
	playerLoginUnSub, err := ps2EventsPublisher.AddHandler(playerLogin)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	playerLogout := make(chan ps2events.PlayerLogout)
	playerLogoutUnSub, err := ps2EventsPublisher.AddHandler(playerLogout)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	skillAdded := make(chan ps2events.SkillAdded)
	skillAddedUnSub, err := ps2EventsPublisher.AddHandler(skillAdded)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	vehicleDestroy := make(chan ps2events.VehicleDestroy)
	vehicleDestroyUnSub, err := ps2EventsPublisher.AddHandler(vehicleDestroy)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
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
	worldIds []ps2.WorldId,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	ps2EventsPublisher *publisher.Publisher,
	charactersTrackerPublisher *publisher.Publisher,
) (*characters_tracker.CharactersTracker, error) {
	const op = "startNewCharactersTracker"
	log := infra.OpLogger(ctx, op)
	charactersTracker := characters_tracker.New(log, worldIds, characterLoader, charactersTrackerPublisher)
	return charactersTracker, startCharactersTracker(
		ctx,
		charactersTracker,
		ps2EventsPublisher,
	)
}
