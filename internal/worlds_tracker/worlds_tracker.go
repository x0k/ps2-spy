package worlds_tracker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrWorldNotFound = fmt.Errorf("world not found")

type metagameEvent struct {
	ps2.MetagameEvent
	StartedAt time.Time
}

type facilityState struct {
	FactionId  ps2_factions.Id
	OutfitId   ps2.OutfitId
	CapturedAt time.Time
}

type zoneState struct {
	IsLocked     bool
	Since        time.Time
	ControlledBy ps2_factions.Id
	IsUnstable   bool
	Events       map[ps2.InstanceId]metagameEvent
	Facilities   map[ps2.FacilityId]facilityState
}

// Returns modified state if `locked` status is changed
func (z zoneState) update(
	isLocked bool,
	since time.Time,
	controlledBy ps2_factions.Id,
	isUnstable bool,
) zoneState {
	// Lock state
	if isLocked {
		// Ignore lock of locked zone to keep original data
		if z.IsLocked {
			return z
		}
		z.IsLocked = true
		z.Since = since
		z.ControlledBy = controlledBy
		// Reset to false, cause zone is locked
		z.IsUnstable = false
		// This changes modifies original state
		clear(z.Events)
		clear(z.Facilities)
	} else {
		// Unlock state
		if z.IsLocked {
			z.IsLocked = false
			z.Since = since
			z.ControlledBy = ps2_factions.None
		}
		// This status may be changed during the `unlocked` state
		z.IsUnstable = isUnstable
	}
	return z
}

type WorldMapLoader = loader.Keyed[ps2.WorldId, ps2.WorldMap]

type WorldsTracker struct {
	log                  *logger.Logger
	worldMapLoader       WorldMapLoader
	worldIds             []ps2.WorldId
	mutex                sync.RWMutex
	worlds               map[ps2.WorldId]map[ps2.ZoneId]zoneState
	invalidationInterval time.Duration
	publisher            pubsub.Publisher[Event]
}

func New(
	log *logger.Logger,
	platform ps2_platforms.Platform,
	invalidationInterval time.Duration,
	publisher pubsub.Publisher[Event],
	worldMapLoader WorldMapLoader,
) *WorldsTracker {
	worldIds := ps2.PlatformWorldIds[platform]
	worlds := make(map[ps2.WorldId]map[ps2.ZoneId]zoneState, len(worldIds))
	now := time.Now()
	for _, worldId := range worldIds {
		world := make(map[ps2.ZoneId]zoneState, len(ps2.ZoneIds))
		for _, zoneId := range ps2.ZoneIds {
			world[zoneId] = zoneState{
				Since:        now,
				ControlledBy: ps2_factions.None,
				Events:       make(map[ps2.InstanceId]metagameEvent, 2),
				Facilities:   make(map[ps2.FacilityId]facilityState, ps2.ZoneFacilitiesCount[zoneId]),
			}
		}
		worlds[worldId] = world
	}
	return &WorldsTracker{
		log:                  log,
		worldIds:             worldIds,
		worldMapLoader:       worldMapLoader,
		worlds:               worlds,
		invalidationInterval: invalidationInterval,
		publisher:            publisher,
	}
}

func (w *WorldsTracker) invalidateEvents(now time.Time) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
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

func (w *WorldsTracker) invalidateWorldFacilitiesTask(
	ctx context.Context,
	wg *sync.WaitGroup,
	now time.Time,
	worldMap ps2.WorldMap,
) {
	defer wg.Done()
	w.mutex.Lock()
	defer w.mutex.Unlock()
	world, ok := w.worlds[worldMap.Id]
	if !ok {
		w.log.Error(ctx, "world not found", slog.String("world_id", string(worldMap.Id)))
		return
	}
	for zoneId, zoneMap := range worldMap.Zones {
		if zoneState, ok := world[zoneId]; ok {
			isLocked := true
			isUnstable := false
			controlledBy := zoneState.ControlledBy
			for facilityId, factionId := range zoneMap.Facilities {
				// Some Oshur and Esamir regions have no associated facilities and as such are ignored
				if facilityId == "" {
					continue
				}
				if factionId == ps2_factions.None {
					isUnstable = true
					continue
				}
				if isLocked {
					if controlledBy == ps2_factions.None {
						controlledBy = factionId
					} else if controlledBy != factionId {
						isLocked = false
					}
				}
				if facility, ok := zoneState.Facilities[facilityId]; (ok && facility.FactionId != factionId) || !ok {
					zoneState.Facilities[facilityId] = facilityState{
						FactionId:  factionId,
						CapturedAt: now,
					}
				}
			}
			world[zoneId] = zoneState.update(
				isLocked,
				now,
				controlledBy,
				isUnstable,
			)
		} else {
			w.log.Error(
				ctx, "zone not found",
				slog.String("world_id", string(worldMap.Id)),
				slog.String("zone_id", string(zoneId)),
			)
			continue
		}
	}
}

func (w *WorldsTracker) invalidateWorldFacilities(
	ctx context.Context,
	wg *sync.WaitGroup,
	now time.Time,
	worldId ps2.WorldId,
) {
	log := w.log.With(slog.String("world_id", string(worldId)))
	worldMap, err := w.worldMapLoader(ctx, worldId)
	if err != nil {
		log.Error(ctx, "failed to invalidate world facilities", sl.Err(err))
		return
	}
	wg.Add(1)
	go w.invalidateWorldFacilitiesTask(ctx, wg, now, worldMap)
}

func (w *WorldsTracker) invalidateFacilities(ctx context.Context, wg *sync.WaitGroup, now time.Time) {
	w.log.Debug(ctx, "facilities invalidation started", slog.Int("worlds_count", len(w.worldIds)))
	for _, worldId := range w.worldIds {
		select {
		case <-ctx.Done():
			return
		default:
			w.invalidateWorldFacilities(ctx, wg, now, worldId)
		}
	}
}

func (w *WorldsTracker) Start(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	w.invalidateFacilities(ctx, wg, time.Now())
	ticker := time.NewTicker(w.invalidationInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case now := <-ticker.C:
			w.invalidateEvents(now)
			// To maintain `unstable` status
			w.invalidateFacilities(ctx, wg, now)
		}
	}
}

func (w *WorldsTracker) HandleMetagameEvent(ctx context.Context, event events.MetagameEvent) error {
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

func (w *WorldsTracker) updateFacilityState(event events.FacilityControl) (ps2.OutfitId, error) {
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
		FactionId:  ps2_factions.Id(event.NewFactionID),
		OutfitId:   ps2.OutfitId(event.OutfitID),
		CapturedAt: capturedAt,
	}
	// Unlock zone
	if zone.IsLocked {
		world[zoneId] = zone.update(
			false,
			capturedAt,
			ps2_factions.None,
			zone.IsUnstable,
		)
	}
	return oldOutfitId, nil
}

func (w *WorldsTracker) HandleFacilityControl(ctx context.Context, event events.FacilityControl) error {
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
	w.publisher.Publish(FacilityControl{
		FacilityControl: event,
		OldOutfitId:     oldOutfitId,
	})
	w.publisher.Publish(FacilityLoss{
		FacilityControl: event,
		OldOutfitId:     oldOutfitId,
	})
	return nil
}

func (w *WorldsTracker) HandleContinentLock(ctx context.Context, event events.ContinentLock) error {
	const op = "worlds_tracker.WorldsTracker.HandleContinentLock"
	w.mutex.Lock()
	defer w.mutex.Unlock()
	world, ok := w.worlds[ps2.WorldId(event.WorldID)]
	if !ok {
		return fmt.Errorf("%s world %q: %w", op, event.WorldID, ErrWorldNotFound)
	}
	zoneId := ps2.ZoneId(event.ZoneID)
	zone, ok := world[zoneId]
	// Non interesting zone
	if !ok {
		return nil
	}
	var since time.Time
	if timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64); err == nil {
		since = time.Unix(timestamp, 0)
	} else {
		since = time.Now()
	}
	// Lock zone
	world[zoneId] = zone.update(
		true,
		since,
		ps2_factions.Id(event.TriggeringFaction),
		false,
	)
	return nil
}

func zoneTerritoryControl(facilities map[ps2.FacilityId]facilityState) ps2.StatPerFactions {
	stat := ps2.StatPerFactions{}
	for _, facility := range facilities {
		stat.All++
		switch facility.FactionId {
		case ps2_factions.NC:
			stat.NC++
		case ps2_factions.TR:
			stat.TR++
		case ps2_factions.VS:
			stat.VS++
		case ps2_factions.NSO:
			stat.NS++
		case ps2_factions.None:
			stat.Other++
		}
	}
	return stat
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
					TerritoryControl: zoneTerritoryControl(zone.Facilities),
				})
			}
		}
	}
	return alerts
}

func (w *WorldsTracker) WorldTerritoryControl(
	ctx context.Context,
	worldId ps2.WorldId,
) (ps2.WorldTerritoryControl, error) {
	const op = "worlds_tracker.WorldsTracker.WorldTerritoryControl"
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	world, ok := w.worlds[worldId]
	if !ok {
		return ps2.WorldTerritoryControl{}, fmt.Errorf("%s world %q: %w", op, worldId, ErrWorldNotFound)
	}
	zones := make([]ps2.ZoneTerritoryControl, 0, len(world))
	for _, zoneId := range ps2.ZoneIds {
		zone, ok := world[zoneId]
		if !ok {
			w.log.Warn(ctx, "zone not found", slog.String("world_id", string(worldId)), slog.String("zone_id", string(zoneId)))
			continue
		}
		zones = append(zones, ps2.ZoneTerritoryControl{
			Id:              zoneId,
			IsOpen:          !zone.IsLocked,
			Since:           zone.Since,
			ControlledBy:    zone.ControlledBy,
			IsStable:        !zone.IsUnstable,
			HasAlerts:       len(zone.Events) > 0,
			StatPerFactions: zoneTerritoryControl(zone.Facilities),
		})
	}
	return ps2.WorldTerritoryControl{
		Id:    worldId,
		Zones: zones,
	}, nil
}
