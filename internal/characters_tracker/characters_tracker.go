package characters_tracker

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
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
)

var ErrWorldPopulationTrackerNotFound = fmt.Errorf("world population tracker not found")

type player struct {
	characterId ps2.CharacterId
	worldId     ps2.WorldId
}

type CharactersTracker struct {
	log                      *slog.Logger
	mutex                    sync.RWMutex
	worldPopulationTrackers  map[ps2.WorldId]worldPopulationTracker
	onlineCharactersTracker  onlineCharactersTracker
	activePlayers            *containers.ExpirationQueue[player]
	inactivityCheckInterval  time.Duration
	inactiveTimeout          time.Duration
	retryableCharacterLoader *retryable.WithArg[ps2.CharacterId, ps2.Character]
}

func New(log *slog.Logger, worldIds []ps2.WorldId, characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character]) *CharactersTracker {
	trackers := make(map[ps2.WorldId]worldPopulationTracker, len(ps2.WorldNames))
	for _, worldId := range worldIds {
		trackers[worldId] = newWorldPopulationTracker()
	}
	return &CharactersTracker{
		log: log.With(
			slog.String("component", "characters_tracker.CharactersTracker"),
			slog.String("world_ids", fmt.Sprintf("%v", worldIds)),
		),
		worldPopulationTrackers: trackers,
		onlineCharactersTracker: newOnlineCharactersTracker(),
		activePlayers:           containers.NewExpirationQueue[player](),
		inactivityCheckInterval: time.Minute,
		inactiveTimeout:         10 * time.Minute,
		retryableCharacterLoader: retryable.NewWithArg[ps2.CharacterId, ps2.Character](
			characterLoader.Load,
		),
	}
}

func (p *CharactersTracker) handleInactive(now time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	count := p.activePlayers.RemoveExpired(now.Add(-p.inactiveTimeout), func(pl player) {
		if w, ok := p.worldPopulationTrackers[pl.worldId]; ok {
			w.HandleInactive(pl.characterId)
		} else {
			p.log.Warn("world not found", slog.String("world_id", string(pl.worldId)))
		}
		p.onlineCharactersTracker.HandleInactive(pl.characterId)
	})
	if count > 0 {
		p.log.Debug("inactive players removed", slog.Int("queue_size", p.activePlayers.Len()), slog.Int("count", count))
	}
}

func (p *CharactersTracker) Start(ctx context.Context) {
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
				p.handleInactive(t)
			}
		}
	}()
}

func (p *CharactersTracker) handleLogin(char ps2.Character) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Push(player{char.Id, char.WorldId})
	if w, ok := p.worldPopulationTrackers[char.WorldId]; ok {
		w.HandleLogin(char)
	} else {
		p.log.Warn("world not found", slog.String("world_id", string(char.WorldId)))
	}
	p.onlineCharactersTracker.HandleLogin(char)
}

func (p *CharactersTracker) HandleLoginTask(ctx context.Context, wg *sync.WaitGroup, event ps2events.PlayerLogin) {
	defer wg.Done()
	charId := ps2.CharacterId(event.CharacterID)
	char, err := p.retryableCharacterLoader.Run(
		ctx,
		charId,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		perform.Log(
			p.log,
			slog.LevelDebug,
			"[ERROR] failed to get character, retrying",
			slog.String("character_id", string(charId)),
		),
	)
	if err != nil {
		p.log.Error("failed to get character", slog.String("character_id", string(charId)), sl.Err(err))
		return
	}
	p.handleLogin(char)
}

func (p *CharactersTracker) HandleLogout(ctx context.Context, event ps2events.PlayerLogout) {
	worldId := ps2.WorldId(event.WorldID)
	charId := ps2.CharacterId(event.CharacterID)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Remove(player{charId, worldId})
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		w.HandleInactive(charId)
	} else {
		p.log.Warn("world not found", slog.String("world_id", string(worldId)))
	}
	p.onlineCharactersTracker.HandleInactive(charId)
}

func (p *CharactersTracker) HandleWorldZoneAction(ctx context.Context, worldId, zoneId, charId string) {
	cId := ps2.CharacterId(charId)
	wId := ps2.WorldId(worldId)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.activePlayers.Push(player{cId, wId})
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneAction(cId, zoneId)
	} else {
		p.log.Warn("world not found", slog.String("world_id", worldId))
	}
}

func (p *CharactersTracker) TrackableOnlineEntities(settings meta.SubscriptionSettings) meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character] {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.onlineCharactersTracker.TrackableOnlineEntities(settings)
}

func (p *CharactersTracker) WorldsPopulation() ps2.WorldsPopulation {
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

func (p *CharactersTracker) DetailedWorldPopulation(worldId ps2.WorldId) (ps2.DetailedWorldPopulation, error) {
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
