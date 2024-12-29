package ps2_census_characters_repo

import (
	"context"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (l *Repository) characterNamesUrl(ns string, charIds []census2.Str) string {
	l.characterNamesMu.Lock()
	defer l.characterNamesMu.Unlock()
	l.characterNamesQuery.SetNamespace(ns)
	l.characterNamesOperand.Set(census2.NewList(charIds, ","))
	return l.client.ToURL(l.characterNamesQuery)
}

func (l *Repository) CharacterNamesByIds(
	ctx context.Context, platform ps2_platforms.Platform, characterIds []ps2.CharacterId,
) (map[ps2.CharacterId]string, error) {
	if len(characterIds) == 0 {
		return nil, nil
	}
	strCharIds := make([]census2.Str, len(characterIds))
	for i, charId := range characterIds {
		strCharIds[i] = census2.Str(string(charId))
	}
	url := l.characterNamesUrl(ps2_platforms.PlatformNamespace(platform), strCharIds)
	chars, err := census2_adapters.RetryableExecutePreparedAndDecode[ps2_collections.CharacterItem](
		ctx,
		l.log,
		l.client,
		ps2_collections.Character,
		url,
	)
	if err != nil {
		return nil, err
	}
	names := make(map[ps2.CharacterId]string, len(chars))
	for _, char := range chars {
		names[ps2.CharacterId(char.CharacterId)] = char.Name.First
	}
	return names, nil
}
