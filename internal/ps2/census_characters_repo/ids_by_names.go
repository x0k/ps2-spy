package census_characters_repo

import (
	"context"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (r *Repository) characterIdsUrl(ns string, values []census2.Str) string {
	r.characterIdsMu.Lock()
	defer r.characterIdsMu.Unlock()
	r.characterIdsOperand.Set(census2.NewList(values, ","))
	r.characterIdsQuery.SetNamespace(ns)
	r.characterIdsQuery.SetLimit(len(values))
	return r.client.ToURL(r.characterIdsQuery)
}

func (r *Repository) CharacterIdsByNames(ctx context.Context, platform ps2_platforms.Platform, characterNames []string) (map[string]ps2.CharacterId, error) {
	if len(characterNames) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(characterNames))
	for i, name := range characterNames {
		values[i] = census2.Str(strings.ToLower(name))
	}
	url := r.characterIdsUrl(ps2_platforms.PlatformNamespace(platform), values)
	chars, err := census2.ExecutePreparedAndDecode[ps2_collections.CharacterItem](
		ctx,
		r.client,
		ps2_collections.Character,
		url,
	)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]ps2.CharacterId, len(chars))
	for _, char := range chars {
		ids[char.Name.First] = ps2.CharacterId(char.CharacterId)
	}
	return ids, nil
}
