package worlds_tracker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
)

var ErrWorldNotFound = fmt.Errorf("world not found")

type metagameEvent struct {
	ps2.MetagameEvent
	StartedAt time.Time
}

type facilityState struct {
	FactionId  factions.Id
	OutfitId   ps2.OutfitId
	CapturedAt time.Time
}

type zoneState struct {
	Events     map[ps2.InstanceId]metagameEvent
	Facilities map[ps2.FacilityId]facilityState
}

type WorldsTracker struct {
	mutex                      sync.RWMutex
	worlds                     map[ps2.WorldId]map[ps2.ZoneId]zoneState
	eventsInvalidationInterval time.Duration
	publisher                  *Publisher
}

func New(eventsInvalidationInterval time.Duration, publisher *Publisher) *WorldsTracker {
	worlds := make(map[ps2.WorldId]map[ps2.ZoneId]zoneState, len(ps2.WorldNames))
	for worldId := range ps2.WorldNames {
		world := make(map[ps2.ZoneId]zoneState, len(ps2.ZoneNames))
		for zoneId := range ps2.ZoneNames {
			world[zoneId] = zoneState{
				Events:     make(map[ps2.InstanceId]metagameEvent, 2),
				Facilities: make(map[ps2.FacilityId]facilityState, ps2.ZoneFacilitiesCount[zoneId]),
			}
		}
		worlds[worldId] = world
	}
	return &WorldsTracker{
		worlds:                     worlds,
		eventsInvalidationInterval: eventsInvalidationInterval,
		publisher:                  publisher,
	}
}

func (w *WorldsTracker) invalidateEvents() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	now := time.Now()
	toRemove := make([]ps2.InstanceId, 0, len(w.worlds))
	for _, world := range w.worlds {
		for _, zone := range world {
			for id, instance := range zone.Events {
				if instance.StartedAt.Add(instance.Duration).Before(now) {
					toRemove = append(toRemove, id)
				}
				for _, id := range toRemove {
					delete(zone.Events, id)
				}
				toRemove = toRemove[:0]
			}
		}
	}
}

func (w *WorldsTracker) Start(ctx context.Context) {
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(w.eventsInvalidationInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.invalidateEvents()
			}
		}
	}()
}

func (w *WorldsTracker) HandleMetagameEvent(ctx context.Context, event ps2events.MetagameEvent) error {
	const op = "worlds_tracker.WorldsTracker.HandleMetagameEvent"
	w.mutex.Lock()
	defer w.mutex.Unlock()
	world, ok := w.worlds[ps2.WorldId(event.WorldID)]
	if !ok {
		return fmt.Errorf("%s world %q: %w", op, event.WorldID, ErrWorldNotFound)
	}
	zone, ok := world[ps2.ZoneId(event.ZoneID)]
	// Non interesting zone
	if !ok {
		return nil
	}
	if event.MetagameEventStateName == ps2.StartedMetagameEventStateName {
		timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("%s invalid timestamp %q: %w", op, event.Timestamp, err)
		}
		e, ok := ps2.MetagameEventsMap[ps2.MetagameEventId(event.MetagameEventID)]
		if !ok {
			return fmt.Errorf("%s unknown metagame event id: %s", op, event.MetagameEventID)
		}
		zone.Events[ps2.InstanceId(event.InstanceID)] = metagameEvent{
			MetagameEvent: e,
			StartedAt:     time.Unix(timestamp, 0),
		}
	} else {
		delete(zone.Events, ps2.InstanceId(event.InstanceID))
	}
	return nil
}

func (w *WorldsTracker) updateFacilityState(event ps2events.FacilityControl) (ps2.OutfitId, error) {
	const op = "worlds_tracker.WorldsTracker.updateFacilityState"
	w.mutex.Lock()
	defer w.mutex.Unlock()
	worldId := ps2.WorldId(event.WorldID)
	world, ok := w.worlds[worldId]
	if !ok {
		return ps2.OutfitId(""), fmt.Errorf("%s world %q: %w", op, worldId, ErrWorldNotFound)
	}
	zoneId := ps2.ZoneId(event.ZoneID)
	zone, ok := world[zoneId]
	// Non interesting zone
	if !ok {
		return ps2.OutfitId(""), nil
	}
	facilityId := ps2.FacilityId(event.FacilityID)
	oldOutfitId := ps2.OutfitId("")
	if facility, ok := zone.Facilities[facilityId]; ok {
		oldOutfitId = facility.OutfitId
	}
	var capturedAt time.Time
	if timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64); err == nil {
		capturedAt = time.Unix(timestamp, 0)
	} else {
		capturedAt = time.Now()
	}
	zone.Facilities[facilityId] = facilityState{
		FactionId:  factions.Id(event.NewFactionID),
		OutfitId:   ps2.OutfitId(event.OutfitID),
		CapturedAt: capturedAt,
	}
	return oldOutfitId, nil
}

func (w *WorldsTracker) HandleFacilityControl(ctx context.Context, event ps2events.FacilityControl) error {
	const op = "worlds_tracker.WorldsTracker.HandleFacilityControl"
	// Defended base
	if event.OldFactionID == event.NewFactionID {
		return nil
	}
	oldOutfitId, err := w.updateFacilityState(event)
	if err != nil {
		return fmt.Errorf("%s failed facility state update: %w", op, err)
	}
	// Event duplication
	if oldOutfitId == ps2.OutfitId(event.OutfitID) && oldOutfitId != "" {
		return nil
	}
	err = w.publisher.Publish(FacilityControl{
		FacilityControl: event,
		OldOutfitId:     oldOutfitId,
	})
	if err != nil {
		return fmt.Errorf("%s failed publishing facility control event: %w", op, err)
	}
	err = w.publisher.Publish(FacilityLoss{
		FacilityControl: event,
		OldOutfitId:     oldOutfitId,
	})
	if err != nil {
		return fmt.Errorf("%s failed publishing facility loss event: %w", op, err)
	}
	return nil
}

func (w *WorldsTracker) HandleContinentLock(ctx context.Context, event ps2events.ContinentLock) error {
	const op = "worlds_tracker.WorldsTracker.HandleContinentLock"
	w.mutex.Lock()
	defer w.mutex.Unlock()
	world, ok := w.worlds[ps2.WorldId(event.WorldID)]
	if !ok {
		return fmt.Errorf("%s world %q: %w", op, event.WorldID, ErrWorldNotFound)
	}
	zone, ok := world[ps2.ZoneId(event.ZoneID)]
	// Non interesting zone
	if !ok {
		return nil
	}
	// TODO: Consider continent lock, check unlocked continents
	clear(zone.Events)
	clear(zone.Facilities)
	return nil
}

func (w *WorldsTracker) Alerts() ps2.Alerts {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	now := time.Now()
	alerts := make(ps2.Alerts, 0, len(w.worlds))
	for worldId, world := range w.worlds {
		for zoneId, zone := range world {
			if len(zone.Events) == 0 {
				continue
			}
			// After continent unlocks, every faction has some facilities,
			// so this calculation is not valid
			// TODO: invalidate facilities states by some ticker
			territoryControl := ps2.StatPerFactions{}
			for _, facility := range zone.Facilities {
				territoryControl.All++
				switch facility.FactionId {
				case factions.NC:
					territoryControl.NC++
				case factions.TR:
					territoryControl.TR++
				case factions.VS:
					territoryControl.VS++
				case factions.NSO:
					territoryControl.NS++
				case factions.None:
					territoryControl.Other++
				}
			}
			for _, event := range zone.Events {
				if event.StartedAt.Add(event.Duration).Before(now) {
					continue
				}
				alerts = append(alerts, ps2.Alert{
					WorldId:          worldId,
					WorldName:        ps2.WorldNameById(worldId),
					ZoneId:           zoneId,
					ZoneName:         ps2.ZoneNameById(zoneId),
					AlertName:        event.Name,
					AlertDescription: event.Description,
					StartedAt:        event.StartedAt,
					Duration:         event.Duration,
					TerritoryControl: territoryControl,
				})
			}
		}
	}
	return alerts
}
