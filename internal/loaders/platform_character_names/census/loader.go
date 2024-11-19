package census_platform_character_names_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Loader struct {
	client  *census2.Client
	queryMu sync.Mutex
	operand *census2.Ptr[census2.List[census2.Str]]
	query   *census2.Query
}

func New(client *census2.Client) *Loader {
	operand := census2.NewPtr(census2.StrList())
	return &Loader{
		client:  client,
		operand: &operand,
		query: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Character).
			Where(census2.Cond("character_id").Equals(&operand)).Show("name.first"),
	}
}

func (l *Loader) toURL(ns string, charIds []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.query.SetNamespace(ns)
	l.operand.Set(census2.NewList(charIds, ","))
	return l.client.ToURL(l.query)
}

func (l *Loader) Load(ctx context.Context, query discord.PlatformQuery[[]ps2.CharacterId]) ([]string, error) {
	if len(query.Value) == 0 {
		return nil, nil
	}
	ns := ps2_platforms.PlatformNamespace(query.Platform)
	strCharIds := make([]census2.Str, len(query.Value))
	for i, charId := range query.Value {
		strCharIds[i] = census2.Str(string(charId))
	}
	url := l.toURL(ns, strCharIds)
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
