package census_outfit_member_ids_loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Loader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand *census2.Ptr[census2.Str]
}

func New(client *census2.Client, namespace string) *Loader {
	operand := census2.NewPtr(census2.Str(""))
	return &Loader{
		client:  client,
		operand: &operand,
		query: census2.NewQuery(census2.GetQuery, namespace, ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&operand)).
			Show("outfit_id").
			WithJoin(
				census2.Join(ps2_collections.OutfitMember).
					Show("character_id").
					InjectAt("outfit_members").
					IsList(true),
			),
	}
}

func (l *Loader) toUrl(outfitId ps2.OutfitId) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(outfitId))
	return l.client.ToURL(l.query)
}

func (l *Loader) Load(ctx context.Context, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
	url := l.toUrl(outfitId)
	outfits, err := census2.ExecutePreparedAndDecode[ps2_collections.OutfitItem](
		ctx,
		l.client,
		ps2_collections.Outfit,
		url,
	)
	if err != nil {
		return nil, err
	}
	if len(outfits) == 0 {
		return nil, fmt.Errorf("outfit %q: %w", string(outfitId), loader.ErrNotFound)
	}
	members := make([]ps2.CharacterId, len(outfits[0].OutfitMembers))
	for i, member := range outfits[0].OutfitMembers {
		members[i] = ps2.CharacterId(member.CharacterId)
	}
	return members, nil
}
