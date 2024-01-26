package population_tracker

import (
	ps2events "github.com/x0k/ps2-spy/internal/lib/census2/streaming/events"
	"github.com/x0k/ps2-spy/internal/meta"
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

func (o *onlineCharactersTracker) TrackableOnlineEntities(settings meta.SubscriptionSettings) meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character] {
	outfits := make(map[ps2.OutfitId][]ps2.Character, len(settings.Outfits))
	for _, outfitId := range settings.Outfits {
		if characters, ok := o.onlineCharacters[outfitId]; ok {
			chars := make([]ps2.Character, 0, len(characters))
			for _, char := range characters {
				chars = append(outfits[outfitId], char)
			}
			outfits[outfitId] = chars
		}
	}
	characters := make([]ps2.Character, 0, len(settings.Characters))
	for _, charId := range settings.Characters {
		if outfitId, ok := o.characterOutfits[charId]; ok {
			characters = append(characters, o.onlineCharacters[outfitId][charId])
		}
	}
	return meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{
		Outfits:    outfits,
		Characters: characters,
	}
}
