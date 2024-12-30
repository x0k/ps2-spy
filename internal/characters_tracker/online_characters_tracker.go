package characters_tracker

import (
	"maps"

	"github.com/x0k/ps2-spy/internal/ps2"
)

type onlineCharactersTracker struct {
	onlineCharactersByOutfit map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character
	characterOutfitMap       map[ps2.CharacterId]ps2.OutfitId
}

func newOnlineCharactersTracker() onlineCharactersTracker {
	return onlineCharactersTracker{
		onlineCharactersByOutfit: make(map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character),
		characterOutfitMap:       make(map[ps2.CharacterId]ps2.OutfitId),
	}
}

// Returns `true` if character was added.
func (o *onlineCharactersTracker) HandleLogin(char ps2.Character) bool {
	oldOutfitId, isOldOutfitExists := o.characterOutfitMap[char.Id]
	if isOldOutfitExists {
		if oldOutfitId == char.OutfitId {
			return false
		}
		delete(o.onlineCharactersByOutfit[oldOutfitId], char.Id)
	}
	outfit, ok := o.onlineCharactersByOutfit[char.OutfitId]
	if !ok {
		outfit = make(map[ps2.CharacterId]ps2.Character)
		o.onlineCharactersByOutfit[char.OutfitId] = outfit
	}
	outfit[char.Id] = char
	o.characterOutfitMap[char.Id] = char.OutfitId
	return !isOldOutfitExists
}

// Returns `true` if character was deleted.
func (o *onlineCharactersTracker) HandleInactive(charId ps2.CharacterId) bool {
	if outfitId, ok := o.characterOutfitMap[charId]; ok {
		delete(o.characterOutfitMap, charId)
		delete(o.onlineCharactersByOutfit[outfitId], charId)
		return true
	}
	return false
}

func (o *onlineCharactersTracker) OutfitMembersOnline(
	outfitIds []ps2.OutfitId,
) map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character {
	outfits := make(map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character, len(outfitIds))
	for _, outfitId := range outfitIds {
		if characters, ok := o.onlineCharactersByOutfit[outfitId]; ok {
			outfits[outfitId] = maps.Clone(characters)
		}
	}
	return outfits
}

func (o *onlineCharactersTracker) CharactersOnline(
	characterIds []ps2.CharacterId,
) map[ps2.CharacterId]ps2.Character {
	characters := make(map[ps2.CharacterId]ps2.Character, len(characterIds))
	for _, charId := range characterIds {
		if outfitId, ok := o.characterOutfitMap[charId]; ok {
			characters[charId] = o.onlineCharactersByOutfit[outfitId][charId]
		}
	}
	return characters
}

func (o *onlineCharactersTracker) isOnline(charId ps2.CharacterId) bool {
	_, ok := o.characterOutfitMap[charId]
	return ok
}
