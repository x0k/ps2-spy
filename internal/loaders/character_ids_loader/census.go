package character_ids_loader

import (
	"context"
	"strings"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand census2.Ptr[census2.List[census2.Str]]
}

func NewCensus(client *census2.Client, ns string) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, ns, collections.Character).
			Where(census2.Cond("name.first_lower").Equals(operand)).Show("character_id"),
	}
}

func (l *CensusLoader) toURL(values []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.NewList(values, ","))
	l.query.SetLimit(len(values))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, charNames []string) ([]string, error) {
	if len(charNames) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(charNames))
	for i, name := range charNames {
		values[i] = census2.Str(strings.ToLower(name))
	}
	url := l.toURL(values)
	chars, err := census2.ExecutePreparedAndDecode[collections.CharacterItem](
		ctx,
		l.client,
		collections.Character,
		url,
	)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(chars))
	for i, char := range chars {
		ids[i] = char.CharacterId
	}
	return ids, nil
}
