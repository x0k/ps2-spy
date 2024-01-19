package outfit_tags_loader

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
		query: census2.NewQuery(census2.GetQuery, ns, collections.Outfit).
			Where(census2.Cond("alias_lower").Equals(operand)).Show("alias"),
	}
}

func (l *CensusLoader) toUrl(values []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.NewList(values, ","))
	l.query.SetLimit(len(values))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, outfitTags []string) ([]string, error) {
	if len(outfitTags) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(outfitTags))
	for i, tag := range outfitTags {
		values[i] = census2.Str(strings.ToLower(tag))
	}
	url := l.toUrl(values)
	outfits, err := census2.ExecutePreparedAndDecode[collections.OutfitItem](
		ctx,
		l.client,
		collections.Outfit,
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
