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
	"github.com/x0k/ps2-spy/internal/ps2"
)

type outfitsOnlineMembersTracker struct {
	characterLoader            loaders.KeyedLoader[ps2.CharacterId, ps2.Character]
	mu                         sync.Mutex
	onlineOutfitMembers        map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character
	characterOutfits           map[ps2.CharacterId]ps2.OutfitId
	missUnregisteredCharacters *expirable.LRU[ps2.CharacterId, struct{}]
}

func newOutfitsOnlineMembersTracker(characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character]) *outfitsOnlineMembersTracker {
	return &outfitsOnlineMembersTracker{
		characterLoader:            characterLoader,
		onlineOutfitMembers:        make(map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character),
		characterOutfits:           make(map[ps2.CharacterId]ps2.OutfitId),
		missUnregisteredCharacters: expirable.NewLRU[ps2.CharacterId, struct{}](0, nil, 5*time.Minute),
	}
}

func (o *outfitsOnlineMembersTracker) registerCharacter(char ps2.Character) {
	o.mu.Lock()
	defer o.mu.Unlock()
	// Player left before character info is loaded.
	// Since logout event is delayed, this condition
	// will be triggered really rarely
	if o.missUnregisteredCharacters.Contains(char.Id) {
		o.missUnregisteredCharacters.Remove(char.Id)
		return
	}
	outfit, ok := o.onlineOutfitMembers[char.OutfitId]
	if !ok {
		outfit = make(map[ps2.CharacterId]ps2.Character)
		o.onlineOutfitMembers[char.OutfitId] = outfit
	}
	outfit[char.Id] = char
	o.characterOutfits[char.Id] = char.OutfitId
}

func (o *outfitsOnlineMembersTracker) unregisterCharacter(charId ps2.CharacterId) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if outfitId, ok := o.characterOutfits[charId]; ok {
		delete(o.characterOutfits, charId)
		delete(o.onlineOutfitMembers[outfitId], charId)
	} else {
		o.missUnregisteredCharacters.Add(charId, struct{}{})
	}
}

func (o *outfitsOnlineMembersTracker) HandleLoginTask(ctx context.Context, wg *sync.WaitGroup, event ps2events.PlayerLogin) {
	defer wg.Done()
	const op = "population_tracker.outfitsOnlineMembersTracker.HandleLogin"
	log := infra.OpLogger(ctx, op)
	charId := ps2.CharacterId(event.CharacterID)
	char, err := o.characterLoader.Load(ctx, charId)
	if err != nil {
		log.Error("failed to get character", slog.String("character_id", string(charId)), sl.Err(err))
		return
	}
	o.registerCharacter(char)
}

func (o *outfitsOnlineMembersTracker) HandleLogout(event ps2events.PlayerLogout) {
	o.unregisterCharacter(ps2.CharacterId(event.CharacterID))
}
