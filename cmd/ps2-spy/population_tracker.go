package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/population_tracker"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func startPopulationTracker(
	ctx context.Context,
	populationTracker *population_tracker.PopulationTracker,
	ps2EventsPublisher *publisher.Publisher,
) error {
	const op = "startPopulationTracker"
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
		defer func() {
			wg.Done()
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
				go populationTracker.HandleLoginTask(ctx, wg, e)
			case e := <-playerLogout:
				populationTracker.HandleLogout(ctx, e)
			case e := <-achievementEarned:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-battleRankUp:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-death:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-gainExperience:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-itemAdded:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-playerFacilityCapture:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-playerFacilityDefend:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-skillAdded:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			case e := <-vehicleDestroy:
				populationTracker.HandleWorldZoneIdAction(ctx, e.WorldID, e.ZoneID, e.CharacterID)
			}
		}
	}()
	return nil
}

func startNewPopulationTracker(
	ctx context.Context,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	ps2EventsPublisher *publisher.Publisher,
) (*population_tracker.PopulationTracker, error) {
	populationTracker := population_tracker.New(characterLoader)
	return populationTracker, startPopulationTracker(
		ctx,
		populationTracker,
		ps2EventsPublisher,
	)
}
