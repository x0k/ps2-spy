package characters_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	client   *census2.Client
	operand  census2.Ptr[census2.List[census2.Str]]
	queryMu  sync.Mutex
	query    *census2.Query
	platform platforms.Platform
}

func NewCensus(client *census2.Client, platform platforms.Platform) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		client:  client,
		operand: operand,
		// TODO: Use show
		query: census2.NewQuery(census2.GetQuery, platforms.PlatformNamespace(platform), collections.Character).
			Where(census2.Cond("character_id").Equals(operand)).
			// TODO: Use join with show
			Resolve("outfit", "world"),
		platform: platform,
	}
}

func (l *CensusLoader) toUrl(charIds []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.NewList(charIds, ","))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, charIds []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
	strCharIds := make([]census2.Str, len(charIds))
	for i, charId := range charIds {
		strCharIds[i] = census2.Str(charId)
	}
	url := l.toUrl(strCharIds)
	chars, err := census2.ExecutePreparedAndDecode[collections.CharacterItem](ctx, l.client, collections.Character, url)
	if err != nil {
		return nil, err
	}
	m := make(map[ps2.CharacterId]ps2.Character, len(chars))
	for _, char := range chars {
		cId := ps2.CharacterId(char.CharacterId)
		m[cId] = ps2.Character{
			Id:        cId,
			FactionId: factions.Id(char.FactionId),
			Name:      char.Name.First,
			OutfitId:  ps2.OutfitId(char.Outfit.OutfitId),
			OutfitTag: char.Outfit.Alias,
			WorldId:   ps2.WorldId(char.WorldId),
			Platform:  l.platform,
		}
	}
	return m, nil
}
