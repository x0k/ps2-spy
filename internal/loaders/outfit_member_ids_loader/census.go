package outfit_member_ids_loader

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand census2.Ptr[census2.Str]
}

func NewCensus(client *census2.Client, platform string) (*CensusLoader, error) {
	const op = "loaders.outfit_member_ids_loader.NewCensus"
	operand := census2.NewPtr(census2.Str(""))
	ns, err := platforms.PlatformNamespace(platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, ns, collections.Outfit).
			Where(census2.Cond("alias_lower").Equals(operand)).
			Show("outfit_id").
			WithJoin(
				census2.Join(collections.OutfitMember).
					Show("character_id").
					InjectAt("outfit_members").
					IsList(true),
			),
	}, nil
}

func (l *CensusLoader) toUrl(outfitTag string) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(strings.ToLower(outfitTag)))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, outfitTag string) ([]string, error) {
	url := l.toUrl(outfitTag)
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
		return nil, fmt.Errorf("outfit %q: %w", outfitTag, loaders.ErrNotFound)
	}
	members := make([]string, len(outfits[0].OutfitMembers))
	for i, member := range outfits[0].OutfitMembers {
		members[i] = member.CharacterId
	}
	return members, nil
}
