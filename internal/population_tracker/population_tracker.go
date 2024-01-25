package population_tracker

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type PopulationTracker struct {
	worldPopulationTrackers     map[ps2.WorldId]*worldPopulationTracker
	outfitsOnlineMembersTracker *outfitsOnlineMembersTracker
}

func NewPopulationTracker(characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character]) *PopulationTracker {
	trackers := make(map[ps2.WorldId]*worldPopulationTracker, len(ps2.WorldNames))
	for worldId := range ps2.WorldNames {
		trackers[worldId] = newWorldPopulationTracker()
	}
	return &PopulationTracker{
		worldPopulationTrackers:     trackers,
		outfitsOnlineMembersTracker: newOutfitsOnlineMembersTracker(characterLoader),
	}
}

func (p *PopulationTracker) HandleLogin(ctx context.Context, event ps2events.PlayerLogin) {
	const op = "population_tracker.PopulationTracker.HandleLogin"
	log := infra.OpLogger(ctx, op)
	worldId := ps2.WorldId(event.WorldID)
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		w.HandleLogin(event)
	} else {
		log.Warn("world not found", slog.String("world_id", string(worldId)))
	}
	wg := infra.Wg(ctx)
	wg.Add(1)
	go p.outfitsOnlineMembersTracker.HandleLoginTask(ctx, wg, event)
}

func (p *PopulationTracker) HandleLogout(ctx context.Context, event ps2events.PlayerLogout) {
	const op = "population_tracker.PopulationTracker.HandleLogout"
	log := infra.OpLogger(ctx, op)
	worldId := ps2.WorldId(event.WorldID)
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		w.HandleLogout(event)
	} else {
		log.Warn("world not found", slog.String("world_id", string(worldId)))
	}
	p.outfitsOnlineMembersTracker.HandleLogout(event)
}

func (p *PopulationTracker) HandleWorldZoneIdAction(ctx context.Context, worldId, zoneId, charId string) {
	const op = "population_tracker.PopulationTracker.HandleWorldZoneIdAction"
	log := infra.OpLogger(ctx, op)
	wId := ps2.WorldId(worldId)
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneIdAction(zoneId, charId)
	} else {
		log.Warn("world not found", slog.String("world_id", string(wId)))
	}
}
