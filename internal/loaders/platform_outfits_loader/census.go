package platform_outfits_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand census2.Ptr[census2.List[census2.Str]]
}

func NewCensus(client *census2.Client) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(operand)).
			Show("outfit_id", "name", "alias"),
	}
}

func (l *CensusLoader) toUrl(ns string, values []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.query.SetNamespace(ns)
	l.query.SetLimit(len(values))
	l.operand.Set(census2.NewList(values, ","))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, query meta.PlatformQuery[ps2.OutfitId]) (map[ps2.OutfitId]ps2.Outfit, error) {
	if len(query.Items) == 0 {
		return nil, nil
	}
	ns := platforms.PlatformNamespace(query.Platform)
	values := make([]census2.Str, len(query.Items))
	for i, outfitId := range query.Items {
		values[i] = census2.Str(outfitId)
	}
	url := l.toUrl(ns, values)
	outfits, err := census2.ExecutePreparedAndDecode[collections.OutfitItem](
		ctx,
		l.client,
		collections.Outfit,
		url,
	)
	if err != nil {
		return nil, err
	}
	res := make(map[ps2.OutfitId]ps2.Outfit, len(outfits))
	for _, outfit := range outfits {
		outfitId := ps2.OutfitId(outfit.OutfitId)
		res[outfitId] = ps2.Outfit{
			Id:       outfitId,
			Name:     outfit.Name,
			Tag:      outfit.Alias,
			Platform: query.Platform,
		}
	}
	return res, nil
}
