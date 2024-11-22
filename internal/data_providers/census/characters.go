package census_data_provider

import (
	"context"
	"log/slog"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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

func (l *DataProvider) charactersUrl(ns string, charIds []census2.Str) string {
	l.charactersMu.Lock()
	defer l.charactersMu.Unlock()
	l.charactersOperand.Set(census2.NewList(charIds, ","))
	l.charactersQuery.SetNamespace(ns)
	return l.client.ToURL(l.characterIdsQuery)
}

func (l *DataProvider) characters(ctx context.Context, ns string, charIds []ps2.CharacterId) ([]ps2_collections.CharacterItem, error) {
	strCharIds := make([]census2.Str, len(charIds))
	for i, charId := range charIds {
		strCharIds[i] = census2.Str(charId)
	}
	url := l.charactersUrl(ns, strCharIds)
	return l.retryableCharactersLoader.Run(
		ctx,
		url,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		perform.Log(
			l.log.Logger,
			slog.LevelDebug,
			"[ERROR] failed to load characters, retrying",
			slog.String("url", url),
		),
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
