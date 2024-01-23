package facilities_manager

import (
	"context"
	"strconv"
	"sync"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type FacilitiesManager struct {
	stateMu sync.Mutex
	// worldId -> zoneId -> facilityId -> outfitId
	state     map[string]map[string]map[string]ps2.OutfitId
	publisher publisher.Abstract[publisher.Event]
}

func New(
	worldIds []string,
	publisher publisher.Abstract[publisher.Event],
) *FacilitiesManager {
	worlds := make(map[string]map[string]map[string]ps2.OutfitId, len(worldIds))
	for _, worldId := range worldIds {
		world := make(map[string]map[string]ps2.OutfitId, len(ps2.ZoneNames))
		for zoneId := range ps2.ZoneNames {
			world[strconv.Itoa(int(zoneId))] = make(map[string]ps2.OutfitId, ps2.ZoneFacilitiesCount[zoneId])
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
	world := fm.state[event.WorldID]
	// This maps make should never happen, but just in case
	if world == nil {
		world = make(map[string]map[string]ps2.OutfitId)
		fm.state[event.WorldID] = world
	}
	zone := world[event.ZoneID]
	// This maps make should never happen, but just in case
	if zone == nil {
		zone = make(map[string]ps2.OutfitId)
		world[event.ZoneID] = zone
	}
	oldOutfitId := zone[event.FacilityID]
	zone[event.FacilityID] = ps2.OutfitId(event.OutfitID)
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
