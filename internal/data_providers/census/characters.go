package census_data_provider

import (
	"context"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (l *DataProvider) charactersUrl(ns string, charIds []census2.Str) string {
	l.charactersMu.Lock()
	defer l.charactersMu.Unlock()
	l.charactersOperand.Set(census2.NewList(charIds, ","))
	l.charactersQuery.SetNamespace(ns)
	return l.client.ToURL(l.charactersQuery)
}

func (l *DataProvider) characters(ctx context.Context, ns string, charIds []ps2.CharacterId) ([]ps2_collections.CharacterItem, error) {
	strCharIds := make([]census2.Str, len(charIds))
	for i, charId := range charIds {
		strCharIds[i] = census2.Str(charId)
	}
	url := l.charactersUrl(ns, strCharIds)
	return census2_adapters.RetryableExecutePreparedAndDecode[ps2_collections.CharacterItem](
		ctx, l.log, l.client, ps2_collections.Character, url,
	)
}

func (l *DataProvider) makeCharacter(platform ps2_platforms.Platform, char ps2_collections.CharacterItem) ps2.Character {
	return ps2.Character{
		Id:        ps2.CharacterId(char.CharacterId),
		FactionId: ps2_factions.Id(char.FactionId),
		Name:      char.Name.First,
		OutfitId:  ps2.OutfitId(char.OutfitMemberExtended.OutfitId),
		OutfitTag: char.OutfitMemberExtended.Alias,
		WorldId:   ps2.WorldId(char.CharactersWorld.WorldId),
		Platform:  platform,
	}
}

func (l *DataProvider) Characters(ctx context.Context, platform ps2_platforms.Platform, charIds []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
	ns := ps2_platforms.PlatformNamespace(platform)
	chars, err := l.characters(ctx, ns, charIds)
	if err != nil {
		return nil, err
	}
	m := make(map[ps2.CharacterId]ps2.Character, len(charIds))
	for _, char := range chars {
		m[ps2.CharacterId(char.CharacterId)] = l.makeCharacter(platform, char)
	}
	// If there are missing characters, then load them directly,
	// otherwise they will be skipped again.
	diff := len(charIds) - len(chars)
	if diff > 0 && len(chars) > 0 {
		missingCharIds := make([]ps2.CharacterId, 0, diff)
		for _, charId := range charIds {
			if _, ok := m[charId]; !ok {
				missingCharIds = append(missingCharIds, charId)
			}
		}
		if chars, err := l.characters(ctx, ns, missingCharIds); err == nil && len(chars) == 1 {
			for _, char := range chars {
				m[ps2.CharacterId(char.CharacterId)] = l.makeCharacter(platform, char)
			}
		}
	}
	return m, nil
}
