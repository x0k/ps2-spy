package facilities_manager

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type FacilitiesManager struct {
	stateMu   sync.Mutex
	state     map[ps2.WorldId]map[ps2.ZoneId]map[ps2.FacilityId]ps2.OutfitId
	publisher publisher.Abstract[publisher.Event]
}

func New(
	worldIds []ps2.WorldId,
	publisher publisher.Abstract[publisher.Event],
) *FacilitiesManager {
	worlds := make(map[ps2.WorldId]map[ps2.ZoneId]map[ps2.FacilityId]ps2.OutfitId, len(worldIds))
	for _, worldId := range worldIds {
		world := make(map[ps2.ZoneId]map[ps2.FacilityId]ps2.OutfitId, len(ps2.ZoneNames))
		for zoneId := range ps2.ZoneNames {
			world[zoneId] = make(map[ps2.FacilityId]ps2.OutfitId, ps2.ZoneFacilitiesCount[zoneId])
		}
		worlds[worldId] = world
	}
	return &FacilitiesManager{
		publisher: publisher,
		state:     worlds,
	}
}

func (fm *FacilitiesManager) updateState(event ps2events.FacilityControl) ps2.OutfitId {
	fm.stateMu.Lock()
	defer fm.stateMu.Unlock()
	wId := ps2.WorldId(event.WorldID)
	world := fm.state[wId]
	// This maps make should never happen, but just in case
	if world == nil {
		world = make(map[ps2.ZoneId]map[ps2.FacilityId]ps2.OutfitId)
		fm.state[wId] = world
	}
	zId := ps2.ZoneId(event.ZoneID)
	zone := world[zId]
	// This maps make should never happen, but just in case
	if zone == nil {
		zone = make(map[ps2.FacilityId]ps2.OutfitId)
		world[zId] = zone
	}
	fId := ps2.FacilityId(event.FacilityID)
	oldOutfitId := zone[fId]
	zone[fId] = ps2.OutfitId(event.OutfitID)
	return oldOutfitId
}

func (fm *FacilitiesManager) FacilityControl(ctx context.Context, event ps2events.FacilityControl) error {
	const op = "facilities_manager.FacilitiesManager.FacilityControl"
	log := infra.OpLogger(ctx, op)
	// Defended base
	if event.OldFactionID == event.NewFactionID {
		return nil
	}
	oldOutfitId := fm.updateState(event)
	err := fm.publisher.Publish(FacilityControl{
		FacilityControl: event,
		OldOutfitId:     oldOutfitId,
	})
	if err != nil {
		log.Error("publishing control event", sl.Err(err))
	}
	err = fm.publisher.Publish(FacilityLoss{
		FacilityControl: event,
		OldOutfitId:     oldOutfitId,
	})
	if err != nil {
		log.Error("publishing loss event", sl.Err(err))
	}
	return nil
}
