package census_data_provider

import (
	"context"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (l *DataProvider) characterIdsUrl(ns string, values []census2.Str) string {
	l.characterIdsMu.Lock()
	defer l.characterIdsMu.Unlock()
	l.characterIdsOperand.Set(census2.NewList(values, ","))
	l.characterIdsQuery.SetNamespace(ns)
	l.characterIdsQuery.SetLimit(len(values))
	return l.client.ToURL(l.characterIdsQuery)
}

func (l *DataProvider) CharacterIds(ctx context.Context, ns string, characterNames []string) ([]ps2.CharacterId, error) {
	if len(characterNames) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(characterNames))
	for i, name := range characterNames {
		values[i] = census2.Str(strings.ToLower(name))
	}
	url := l.characterIdsUrl(ns, values)
	chars, err := census2.ExecutePreparedAndDecode[ps2_collections.CharacterItem](
		ctx,
		l.client,
		ps2_collections.Character,
		url,
	)
	if err != nil {
		return nil, err
	}
	ids := make([]ps2.CharacterId, len(chars))
	for i, char := range chars {
		ids[i] = ps2.CharacterId(char.CharacterId)
	}
	return ids, nil
}

func (l *DataProvider) characterNamesUrl(ns string, charIds []census2.Str) string {
	l.characterNamesMu.Lock()
	defer l.characterNamesMu.Unlock()
	l.characterNamesQuery.SetNamespace(ns)
	l.characterNamesOperand.Set(census2.NewList(charIds, ","))
	return l.client.ToURL(l.characterNamesQuery)
}

func (l *DataProvider) CharacterNames(ctx context.Context, ns string, characterIds []ps2.CharacterId) ([]string, error) {
	if len(characterIds) == 0 {
		return nil, nil
	}
	strCharIds := make([]census2.Str, len(characterIds))
	for i, charId := range characterIds {
		strCharIds[i] = census2.Str(string(charId))
	}
	url := l.characterNamesUrl(ns, strCharIds)
	chars, err := census2.ExecutePreparedAndDecode[ps2_collections.CharacterItem](
		ctx,
		l.client,
		ps2_collections.Character,
		url,
	)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(chars))
	for i, char := range chars {
		names[i] = char.Name.First
	}
	return names, nil
}
