package census_outfits_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Loader struct {
	client   *census2.Client
	queryMu  sync.Mutex
	query    *census2.Query
	operand  *census2.Ptr[census2.List[census2.Str]]
	platform ps2_platforms.Platform
}

func New(client *census2.Client, platform ps2_platforms.Platform) *Loader {
	operand := census2.NewPtr(census2.StrList())
	return &Loader{
		client:  client,
		operand: &operand,
		query: census2.NewQuery(census2.GetQuery, ps2_platforms.PlatformNamespace(platform), ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&operand)).
			Show("outfit_id", "name", "alias"),
		platform: platform,
	}
}

func (l *Loader) toUrl(values []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.query.SetLimit(len(values))
	l.operand.Set(census2.NewList(values, ","))
	return l.client.ToURL(l.query)
}

func (l *Loader) Load(ctx context.Context, outfitIds []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, error) {
	if len(outfitIds) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(outfitIds))
	for i, outfitId := range outfitIds {
		values[i] = census2.Str(outfitId)
	}
	url := l.toUrl(values)
	outfits, err := census2.ExecutePreparedAndDecode[ps2_collections.OutfitItem](
		ctx,
		l.client,
		ps2_collections.Outfit,
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
			Platform: l.platform,
		}
	}
	return res, nil
}
