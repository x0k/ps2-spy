package outfit_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand census2.Ptr[census2.Str]
}

func NewCensus(client *census2.Client, ns string) *CensusLoader {
	operand := census2.NewPtr(census2.Str(""))
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, ns, collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(operand)).
			Show("outfit_id", "name", "alias"),
	}
}

func (l *CensusLoader) toUrl(outfitId ps2.OutfitId) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(outfitId))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, outfitId ps2.OutfitId) (ps2.Outfit, error) {
	url := l.toUrl(outfitId)
	outfits, err := census2.ExecutePreparedAndDecode[collections.OutfitItem](
		ctx,
		l.client,
		collections.Outfit,
		url,
	)
	if err != nil {
		return ps2.Outfit{}, err
	}
	if len(outfits) == 0 {
		return ps2.Outfit{}, loaders.ErrNotFound
	}
	return ps2.Outfit{
		Id:   ps2.OutfitId(outfits[0].OutfitId),
		Name: outfits[0].Name,
		Tag:  outfits[0].Alias,
	}, nil
}
