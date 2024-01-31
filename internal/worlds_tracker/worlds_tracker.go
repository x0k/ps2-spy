package worlds_tracker

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type metagameEvent struct {
	ps2.MetagameEvent
	StartedAt time.Time
}

type WorldsTracker struct {
	log                        *slog.Logger
	mutex                      sync.RWMutex
	worlds                     map[ps2.WorldId]map[ps2.ZoneId]map[ps2.InstanceId]metagameEvent
	eventsInvalidationInterval time.Duration
}

func New(log *slog.Logger, eventsInvalidationInterval time.Duration) *WorldsTracker {
	worlds := make(map[ps2.WorldId]map[ps2.ZoneId]map[ps2.InstanceId]metagameEvent, len(ps2.WorldNames))
	for worldId := range ps2.WorldNames {
		world := make(map[ps2.ZoneId]map[ps2.InstanceId]metagameEvent, len(ps2.ZoneNames))
		for zoneId := range ps2.ZoneNames {
			world[zoneId] = make(map[ps2.InstanceId]metagameEvent)
		}
		worlds[worldId] = world
	}
	return &WorldsTracker{
		log:                        log.With(slog.String("component", "worlds_tracker.WorldsTracker")),
		worlds:                     worlds,
		eventsInvalidationInterval: eventsInvalidationInterval,
	}
}

func (w *WorldsTracker) invalidateEvents() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	now := time.Now()
	toRemove := make([]ps2.InstanceId, 0, len(w.worlds))
	for _, world := range w.worlds {
		for _, zone := range world {
			for id, instance := range zone {
				if instance.StartedAt.Add(instance.Duration).Before(now) {
					toRemove = append(toRemove, id)
				}
				for _, id := range toRemove {
					delete(zone, id)
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

func (w *WorldsTracker) HandleMetagameEvent(ctx context.Context, event ps2events.MetagameEvent) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	world, ok := w.worlds[ps2.WorldId(event.WorldID)]
	if !ok {
		w.log.LogAttrs(ctx, slog.LevelError, "world not found", slog.String("world_id", string(event.WorldID)))
		return
	}
	zone, ok := world[ps2.ZoneId(event.ZoneID)]
	// Non interesting zone
	if !ok {
		return
	}
	if event.MetagameEventStateName == ps2.StartedMetagameEventStateName {
		timestamp, err := strconv.ParseInt(event.Timestamp, 10, 64)
		if err != nil {
			w.log.LogAttrs(ctx, slog.LevelError, "failed to parse timestamp", slog.String("timestamp", event.Timestamp))
			return
		}
		e, ok := ps2.MetagameEventsMap[ps2.MetagameEventId(event.MetagameEventID)]
		if !ok {
			w.log.LogAttrs(ctx, slog.LevelError, "metagame event not found", slog.String("metagame_event_id", event.MetagameEventID))
			return
		}
		zone[ps2.InstanceId(event.InstanceID)] = metagameEvent{
			MetagameEvent: e,
			StartedAt:     time.Unix(timestamp, 0),
		}
	} else {
		delete(zone, ps2.InstanceId(event.InstanceID))
	}

}

func (w *WorldsTracker) Alerts() ps2.Alerts {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	now := time.Now()
	alerts := make(ps2.Alerts, 0, len(w.worlds))
	for worldId, world := range w.worlds {
		for zoneId, zone := range world {
			for _, event := range zone {
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
					// TODO: add territory control
					TerritoryControl: ps2.StatsByFactions{},
				})
			}
		}
	}
	return alerts
}
