package population_tracker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
)

var ErrWorldPopulationTrackerNotFound = fmt.Errorf("world population tracker not found")

// TODO: Track characters activity to remove inactive characters
type PopulationTracker struct {
	characterLoader         loaders.KeyedLoader[ps2.CharacterId, ps2.Character]
	mutex                   *sync.RWMutex
	worldPopulationTrackers map[ps2.WorldId]*worldPopulationTracker
	onlineCharactersTracker *onlineCharactersTracker
	unhandledLeftCharacters *expirable.LRU[ps2.CharacterId, struct{}]
}

func New(ctx context.Context, characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character]) *PopulationTracker {
	trackers := make(map[ps2.WorldId]*worldPopulationTracker, len(ps2.WorldNames))
	for worldId := range ps2.WorldNames {
		trackers[worldId] = newWorldPopulationTracker()
	}
	onlineCharactersTracker := newOnlineCharactersTracker()
	mutex := &sync.RWMutex{}
	return &PopulationTracker{
		mutex:                   mutex,
		characterLoader:         characterLoader,
		unhandledLeftCharacters: expirable.NewLRU[ps2.CharacterId, struct{}](0, nil, 5*time.Minute),
		worldPopulationTrackers: trackers,
		onlineCharactersTracker: onlineCharactersTracker,
	}
}

func (p *PopulationTracker) handleLogin(log *slog.Logger, char ps2.Character) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// Player left before character info is loaded.
	// Since logout event is delayed, this condition
	// will be triggered really rarely
	if p.unhandledLeftCharacters.Contains(char.Id) {
		log.Warn("character is already logged out", slog.String("character_id", string(char.Id)))
		p.unhandledLeftCharacters.Remove(char.Id)
		return

	}
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
			if err != nil {
				log.Warn("failed to get character, retrying", slog.String("character_id", string(charId)), sl.Err(err))
			}
			return err
		},
		While: retry.ContextIsNotCanceledAndMaxRetriesNotExceeded(3),
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
	handled := true
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		handled = handled && w.HandleLogout(event)
	} else {
		log.Warn("world not found", slog.String("world_id", string(worldId)))
	}
	handled = handled && p.onlineCharactersTracker.HandleLogout(event)
	if !handled {
		p.unhandledLeftCharacters.Add(charId, struct{}{})
	}
}

func (p *PopulationTracker) HandleWorldZoneIdAction(ctx context.Context, worldId, zoneId, charId string) {
	const op = "population_tracker.PopulationTracker.HandleWorldZoneIdAction"
	log := infra.OpLogger(ctx, op)
	wId := ps2.WorldId(worldId)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneIdAction(log, zoneId, charId)
	} else {
		log.Warn("world not found", slog.String("world_id", string(wId)))
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
	fmt.Println("zonesPopulation", zonesPopulation)
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
