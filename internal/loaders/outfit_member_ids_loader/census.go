package outfit_member_ids_loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand census2.Ptr[census2.Str]
}

func NewCensus(client *census2.Client, namespace string) *CensusLoader {
	const op = "loaders.outfit_member_ids_loader.NewCensus"
	operand := census2.NewPtr(census2.Str(""))
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, namespace, collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(operand)).
			Show("outfit_id").
			WithJoin(
				census2.Join(collections.OutfitMember).
					Show("character_id").
					InjectAt("outfit_members").
					IsList(true),
			),
	}
}

func (l *CensusLoader) toUrl(outfitId ps2.OutfitId) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(outfitId))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
	url := l.toUrl(outfitId)
	outfits, err := census2.ExecutePreparedAndDecode[collections.OutfitItem](
		ctx,
		l.client,
		collections.Outfit,
		url,
	)
	if err != nil {
		return nil, err
	}
	if len(outfits) == 0 {
		return nil, fmt.Errorf("outfit %q: %w", string(outfitId), loaders.ErrNotFound)
	}
	members := make([]ps2.CharacterId, len(outfits[0].OutfitMembers))
	for i, member := range outfits[0].OutfitMembers {
		members[i] = ps2.CharacterId(member.CharacterId)
	}
	return members, nil
}
