package characters_loader

import (
	"context"
	"strconv"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	client  *census2.Client
	operand census2.Ptr[census2.List[census2.Str]]
	queryMu sync.Mutex
	query   *census2.Query
}

func NewCensus(client *census2.Client, namespace string) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, namespace, collections.Character).
			Where(census2.Cond("character_id").Equals(operand)).
			Resolve("outfit", "world"),
	}
}

func (l *CensusLoader) toUrl(charIds []string) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.StrList(charIds...))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, charIds []string) (map[string]ps2.Character, error) {
	url := l.toUrl(charIds)
	chars, err := census2.ExecutePreparedAndDecode[collections.CharacterItem](ctx, l.client, collections.Character, url)
	if err != nil {
		return nil, err
	}
	m := make(map[string]ps2.Character, len(chars))
	for _, char := range chars {
		wId, err := strconv.Atoi(char.WorldId)
		if err != nil {
			continue
		}
		m[char.CharacterId] = ps2.Character{
			Id:        char.CharacterId,
			FactionId: char.FactionId,
			Name:      char.Name.First,
			OutfitTag: char.Outfit.Alias,
			WorldId:   ps2.WorldId(wId),
			Platform:  platforms.PC,
		}
	}
	return m, nil
}
