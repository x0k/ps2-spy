package census_platform_outfit_tags_loader

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
	query   *census2.Query
	operand *census2.Ptr[census2.List[census2.Str]]
}

func New(client *census2.Client) *Loader {
	operand := census2.NewPtr(census2.StrList())
	return &Loader{
		client:  client,
		operand: &operand,
		query: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&operand)).Show("alias"),
	}
}

func (l *Loader) toUrl(ns string, values []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.query.SetNamespace(ns)
	l.query.SetLimit(len(values))
	l.operand.Set(census2.NewList(values, ","))
	return l.client.ToURL(l.query)
}

func (l *Loader) Load(ctx context.Context, query discord.PlatformQuery[[]ps2.OutfitId]) ([]string, error) {
	if len(query.Value) == 0 {
		return nil, nil
	}
	ns := ps2_platforms.PlatformNamespace(query.Platform)
	values := make([]census2.Str, len(query.Value))
	for i, outfitId := range query.Value {
		values[i] = census2.Str(outfitId)
	}
	url := l.toUrl(ns, values)
	outfits, err := census2.ExecutePreparedAndDecode[ps2_collections.OutfitItem](
		ctx,
		l.client,
		ps2_collections.Outfit,
		url,
	)
	if err != nil {
		return nil, err
	}
	tags := make([]string, len(outfits))
	for i, outfit := range outfits {
		tags[i] = outfit.Alias
	}
	return tags, nil
}
