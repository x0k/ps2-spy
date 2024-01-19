package platform_character_names_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/bot/handlers/command/channel_setup_command_handler"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	operand census2.Ptr[census2.List[census2.Str]]
	query   *census2.Query
}

func NewCensus(client *census2.Client) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Character).
			Where(census2.Cond("character_id").Equals(operand)).Show("name.first"),
	}
}

func (l *CensusLoader) toURL(ns string, charIds []string) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.query.SetNamespace(ns)
	l.operand.Set(census2.StrList(charIds...))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, query channel_setup_command_handler.PlatformQuery) ([]string, error) {
	if len(query.Items) == 0 {
		return nil, nil
	}
	ns, err := platforms.PlatformNamespace(query.Platform)
	if err != nil {
		return nil, err
	}
	url := l.toURL(ns, query.Items)
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
