package census_data_provider

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/shared"
)

func (l *DataProvider) outfitMemberIdsUrl(ns string, outfitId ps2.OutfitId) string {
	l.outfitMemberIdsMu.Lock()
	defer l.outfitMemberIdsMu.Unlock()
	l.outfitMemberIdsOperand.Set(census2.Str(outfitId))
	l.outfitMemberIdsQuery.SetNamespace(ns)
	return l.client.ToURL(l.outfitMemberIdsQuery)
}

func (l *DataProvider) OutfitMemberIds(ctx context.Context, ns string, outfitId ps2.OutfitId) ([]ps2.CharacterId, error) {
	url := l.outfitMemberIdsUrl(ns, outfitId)
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
		return nil, fmt.Errorf("outfit %q: %w", string(outfitId), shared.ErrNotFound)
	}
	members := make([]ps2.CharacterId, len(outfits[0].OutfitMembers))
	for i, member := range outfits[0].OutfitMembers {
		members[i] = ps2.CharacterId(member.CharacterId)
	}
	return members, nil
}
