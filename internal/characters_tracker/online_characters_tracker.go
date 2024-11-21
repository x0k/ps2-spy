package characters_tracker

import (
	"github.com/x0k/ps2-spy/internal/discord"
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

func (o *onlineCharactersTracker) TrackableOnlineEntities(
	settings discord.TrackableEntities[[]ps2.OutfitId, []ps2.CharacterId],
) discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character] {
	outfits := make(map[ps2.OutfitId][]ps2.Character, len(settings.Outfits))
	for _, outfitId := range settings.Outfits {
		if characters, ok := o.onlineCharactersByOutfit[outfitId]; ok {
			chars := make([]ps2.Character, 0, len(characters))
			for _, char := range characters {
				chars = append(chars, char)
			}
			outfits[outfitId] = chars
		}
	}
	characters := make([]ps2.Character, 0, len(settings.Characters))
	for _, charId := range settings.Characters {
		if outfitId, ok := o.characterOutfitMap[charId]; ok {
			characters = append(characters, o.onlineCharactersByOutfit[outfitId][charId])
		}
	}
	return discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{
		Outfits:    outfits,
		Characters: characters,
	}
}
