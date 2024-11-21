package census_character_ids_loader

import (
	"context"
	"strings"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Loader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand *census2.Ptr[census2.List[census2.Str]]
}

func New(client *census2.Client, ns string) *Loader {
	operand := census2.NewPtr(census2.StrList())
	return &Loader{
		client:  client,
		operand: &operand,
		query: census2.NewQuery(census2.GetQuery, ns, ps2_collections.Character).
			Where(census2.Cond("name.first_lower").Equals(&operand)).Show("character_id"),
	}
}

func (l *Loader) toURL(values []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.NewList(values, ","))
	l.query.SetLimit(len(values))
	return l.client.ToURL(l.query)
}

func (l *Loader) Load(ctx context.Context, characterNames []string) ([]ps2.CharacterId, error) {
	if len(characterNames) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(characterNames))
	for i, name := range characterNames {
		values[i] = census2.Str(strings.ToLower(name))
	}
	url := l.toURL(values)
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