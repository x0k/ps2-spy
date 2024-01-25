package population_tracker

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type WorldPopulationTracker struct {
	population           int
	characterLoader      loaders.KeyedLoader[ps2.CharacterId, ps2.Character]
	onlineCharacters     map[ps2.CharacterId]ps2.Character
	onlineOutfitMembers  map[ps2.OutfitId][]ps2.CharacterId
	charactersLastZoneId map[ps2.CharacterId]ps2.ZoneId
	zonesPopulation      map[ps2.ZoneId]int
}

func (w *WorldPopulationTracker) HandleLogin(event ps2events.PlayerLogin) {
	w.population += 1
}

func (w *WorldPopulationTracker) HandleLogout(event ps2events.PlayerLogout) {
	w.population -= 1
}

type PopulationTracker struct {
	worldPopulationTrackers map[string]WorldPopulationTracker
}

func (p *PopulationTracker) HandleLogin(ctx context.Context, event ps2events.PlayerLogin) {
	const op = "population_tracker.PopulationTracker.HandleLogin"
	log := infra.OpLogger(ctx, op)
	if w, ok := p.worldPopulationTrackers[event.WorldID]; ok {
		w.HandleLogin(event)
	} else {
		log.Warn("world not found", slog.String("world_id", event.WorldID))
	}
}

func (p *PopulationTracker) HandleLogout(ctx context.Context, event ps2events.PlayerLogout) {
	const op = "population_tracker.PopulationTracker.HandleLogout"
	log := infra.OpLogger(ctx, op)
	if w, ok := p.worldPopulationTrackers[event.WorldID]; ok {
		w.HandleLogout(event)
	} else {
		log.Warn("world not found", slog.String("world_id", event.WorldID))
	}
}
