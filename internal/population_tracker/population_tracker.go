package population_tracker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
)

var ErrWorldPopulationTrackerNotFound = fmt.Errorf("world population tracker not found")

type player struct {
	characterId ps2.CharacterId
	worldId     ps2.WorldId
}

type PopulationTracker struct {
	characterLoader         loaders.KeyedLoader[ps2.CharacterId, ps2.Character]
	mutex                   sync.RWMutex
	worldPopulationTrackers map[ps2.WorldId]worldPopulationTracker
	onlineCharactersTracker onlineCharactersTracker
	activePlayers           *containers.ExpirationQueue[player]
	inactivityCheckInterval time.Duration
	inactiveTimeout         time.Duration
}

func New(ctx context.Context, worldIds []ps2.WorldId, characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character]) *PopulationTracker {
	trackers := make(map[ps2.WorldId]worldPopulationTracker, len(ps2.WorldNames))
	for _, worldId := range worldIds {
		trackers[worldId] = newWorldPopulationTracker()
	}
	return &PopulationTracker{
		characterLoader:         characterLoader,
		worldPopulationTrackers: trackers,
		onlineCharactersTracker: newOnlineCharactersTracker(),
		activePlayers:           containers.NewExpirationQueue[player](),
		inactivityCheckInterval: time.Minute,
		inactiveTimeout:         10 * time.Minute,
	}
}

func (p *PopulationTracker) handleInactive(log *slog.Logger, now time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	count := p.activePlayers.RemoveExpired(now.Add(-p.inactiveTimeout), func(pl player) {
		if w, ok := p.worldPopulationTrackers[pl.worldId]; ok {
			w.HandleInactive(pl.characterId)
		} else {
			log.Warn("world not found", slog.String("world_id", string(pl.worldId)))
		}
		p.onlineCharactersTracker.HandleInactive(pl.characterId)
	})
	log.Debug("inactive players removed", slog.Int("queue_size", p.activePlayers.Len()), slog.Int("count", count))
}

func (p *PopulationTracker) Start(ctx context.Context) {
	const op = "population_tracker.PopulationTracker.Start"
	log := infra.OpLogger(ctx, op)
	wg := infra.Wg(ctx)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(p.inactivityCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				p.handleInactive(log, t)
			}
		}
	}()
}

func (p *PopulationTracker) handleLogin(log *slog.Logger, char ps2.Character) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Push(player{char.Id, char.WorldId})
	if w, ok := p.worldPopulationTrackers[char.WorldId]; ok {
		w.HandleLogin(char)
	} else {
		log.Warn("world not found", slog.String("world_id", string(char.WorldId)))
	}
	p.onlineCharactersTracker.HandleLogin(char)
}

func (p *PopulationTracker) HandleLoginTask(ctx context.Context, wg *sync.WaitGroup, event ps2events.PlayerLogin) {
	defer wg.Done()
	const op = "population_tracker.PopulationTracker.HandleLogin"
	log := infra.OpLogger(ctx, op)
	charId := ps2.CharacterId(event.CharacterID)
	var char ps2.Character
	var err error
	retry.RetryWhileWithRecover(retry.Retryable{
		Try: func() error {
			char, err = p.characterLoader.Load(ctx, charId)
			return err
		},
		While: retry.ContextIsNotCanceledAndMaxRetriesNotExceeded(3),
		BeforeSleep: func(d time.Duration) {
			log.Debug(
				"[ERROR] failed to get character, retrying",
				slog.Duration("after", d),
				slog.String("character_id", string(charId)),
				sl.Err(err),
			)
		},
	})
	if err != nil {
		log.Error("failed to get character", slog.String("character_id", string(charId)), sl.Err(err))
		return
	}
	p.handleLogin(log, char)
}

func (p *PopulationTracker) HandleLogout(ctx context.Context, event ps2events.PlayerLogout) {
	const op = "population_tracker.PopulationTracker.HandleLogout"
	log := infra.OpLogger(ctx, op)
	worldId := ps2.WorldId(event.WorldID)
	charId := ps2.CharacterId(event.CharacterID)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Remove(player{charId, worldId})
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		w.HandleInactive(charId)
	} else {
		log.Warn("world not found", slog.String("world_id", string(worldId)))
	}
	p.onlineCharactersTracker.HandleInactive(charId)
}

func (p *PopulationTracker) HandleWorldZoneIdAction(ctx context.Context, worldId, zoneId, charId string) {
	const op = "population_tracker.PopulationTracker.HandleWorldZoneIdAction"
	log := infra.OpLogger(ctx, op)
	cId := ps2.CharacterId(charId)
	wId := ps2.WorldId(worldId)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Push(player{cId, wId})
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneIdAction(log, cId, zoneId)
	} else {
		log.Warn("world not found", slog.String("world_id", worldId))
	}
}

func (p *PopulationTracker) TrackableOnlineEntities(settings meta.SubscriptionSettings) meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character] {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.onlineCharactersTracker.TrackableOnlineEntities(settings)
}

func (p *PopulationTracker) WorldsPopulation() ps2.WorldsPopulation {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	total := 0
	worlds := make([]ps2.WorldPopulation, 0, len(p.worldPopulationTrackers))
	for worldId, worldTracker := range p.worldPopulationTrackers {
		worldPopulation := worldTracker.Population()
		other := worldPopulation[factions.None]
		vs := worldPopulation[factions.VS]
		nc := worldPopulation[factions.NC]
		tr := worldPopulation[factions.TR]
		ns := worldPopulation[factions.NSO]
		all := vs + nc + tr + ns + other
		total += all
		worlds = append(worlds, ps2.WorldPopulation{
			Id:   worldId,
			Name: ps2.WorldNames[worldId],
			StatsByFactions: ps2.StatsByFactions{
				All:   all,
				VS:    vs,
				NC:    nc,
				TR:    tr,
				NS:    ns,
				Other: other,
			},
		})
	}
	return ps2.WorldsPopulation{
		Total:  total,
		Worlds: worlds,
	}
}

func (p *PopulationTracker) DetailedWorldPopulation(worldId ps2.WorldId) (ps2.DetailedWorldPopulation, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	tracker, ok := p.worldPopulationTrackers[worldId]
	if !ok {
		return ps2.DetailedWorldPopulation{}, ErrWorldPopulationTrackerNotFound
	}
	zonesPopulation := tracker.ZonesPopulation()
	total := 0
	zones := make([]ps2.ZonePopulation, 0, len(zonesPopulation))
	for zoneId, zonePopulation := range zonesPopulation {
		other := zonePopulation[factions.None]
		vs := zonePopulation[factions.VS]
		nc := zonePopulation[factions.NC]
		tr := zonePopulation[factions.TR]
		ns := zonePopulation[factions.NSO]
		all := vs + nc + tr + ns + other
		total += all
		zones = append(zones, ps2.ZonePopulation{
			Id:   zoneId,
			Name: ps2.ZoneNames[zoneId],
			// TODO: Track this
			IsOpen: all > 0,
			StatsByFactions: ps2.StatsByFactions{
				All:   all,
				VS:    vs,
				NC:    nc,
				TR:    tr,
				NS:    ns,
				Other: other,
			},
		})
	}
	return ps2.DetailedWorldPopulation{
		Id:    worldId,
		Name:  ps2.WorldNameById(worldId),
		Total: total,
		Zones: zones,
	}, nil
}
