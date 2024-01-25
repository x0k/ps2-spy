package population_tracker

import (
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type onlineCharactersTracker struct {
	onlineCharacters map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character
	characterOutfits map[ps2.CharacterId]ps2.OutfitId
}

func newOnlineCharactersTracker() *onlineCharactersTracker {
	return &onlineCharactersTracker{
		onlineCharacters: make(map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character),
		characterOutfits: make(map[ps2.CharacterId]ps2.OutfitId),
	}
}

func (o *onlineCharactersTracker) HandleLogin(char ps2.Character) {
	outfit, ok := o.onlineCharacters[char.OutfitId]
	if !ok {
		outfit = make(map[ps2.CharacterId]ps2.Character)
		o.onlineCharacters[char.OutfitId] = outfit
	}
	outfit[char.Id] = char
	o.characterOutfits[char.Id] = char.OutfitId
}

func (o *onlineCharactersTracker) HandleLogout(event ps2events.PlayerLogout) bool {
	charId := ps2.CharacterId(event.CharacterID)
	if outfitId, ok := o.characterOutfits[charId]; ok {
		delete(o.characterOutfits, charId)
		delete(o.onlineCharacters[outfitId], charId)
		return true
	}
	return false
}
