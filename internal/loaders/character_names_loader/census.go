package character_names_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	operand census2.Ptr[census2.List[census2.Str]]
	query   *census2.Query
}

func NewCensus(client *census2.Client, ns string) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, ns, collections.Character).
			Where(census2.Cond("character_id").Equals(operand)).Show("name.first"),
	}
}

func (l *CensusLoader) toURL(charIds []string) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.StrList(charIds...))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, charIds []string) ([]string, error) {
	if len(charIds) == 0 {
		return nil, nil
	}
	url := l.toURL(charIds)
	chars, err := census2.ExecutePreparedAndDecode[collections.CharacterItem](
		ctx,
		l.client,
		collections.Character,
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
