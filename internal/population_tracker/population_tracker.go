package population_tracker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/x0k/ps2-spy/internal/infra"
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/retry"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type PopulationTracker struct {
	characterLoader             loaders.KeyedLoader[ps2.CharacterId, ps2.Character]
	mu                          sync.Mutex
	worldPopulationTrackers     map[ps2.WorldId]*worldPopulationTracker
	outfitsOnlineMembersTracker *onlineCharactersTracker
	unhandledLeftCharacters     *expirable.LRU[ps2.CharacterId, struct{}]
}

func New(characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character]) *PopulationTracker {
	trackers := make(map[ps2.WorldId]*worldPopulationTracker, len(ps2.WorldNames))
	for worldId := range ps2.WorldNames {
		trackers[worldId] = newWorldPopulationTracker()
	}
	return &PopulationTracker{
		characterLoader:             characterLoader,
		unhandledLeftCharacters:     expirable.NewLRU[ps2.CharacterId, struct{}](0, nil, 5*time.Minute),
		worldPopulationTrackers:     trackers,
		outfitsOnlineMembersTracker: newOnlineCharactersTracker(),
	}
}

func (p *PopulationTracker) handleLogin(log *slog.Logger, char ps2.Character) {
	p.mu.Lock()
	defer p.mu.Unlock()
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
	p.outfitsOnlineMembersTracker.HandleLogin(char)
}

func (p *PopulationTracker) HandleLoginTask(ctx context.Context, wg *sync.WaitGroup, event ps2events.PlayerLogin) {
	const op = "population_tracker.PopulationTracker.HandleLogin"
	defer wg.Done()
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
	handled := true
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.worldPopulationTrackers[worldId]; ok {
		handled = handled && w.HandleLogout(event)
	} else {
		log.Warn("world not found", slog.String("world_id", string(worldId)))
	}
	handled = handled && p.outfitsOnlineMembersTracker.HandleLogout(event)
	if !handled {
		p.unhandledLeftCharacters.Add(ps2.CharacterId(event.CharacterID), struct{}{})
	}
}

func (p *PopulationTracker) HandleWorldZoneIdAction(ctx context.Context, worldId, zoneId, charId string) {
	const op = "population_tracker.PopulationTracker.HandleWorldZoneIdAction"
	log := infra.OpLogger(ctx, op)
	wId := ps2.WorldId(worldId)
	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.worldPopulationTrackers[wId]; ok {
		w.HandleZoneIdAction(zoneId, charId)
	} else {
		log.Warn("world not found", slog.String("world_id", string(wId)))
	}
}
