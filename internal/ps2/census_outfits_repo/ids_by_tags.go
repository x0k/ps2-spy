package census_ps2_outfits_repo

import (
	"context"
	"strings"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (l *Repository) outfitIdsUrl(ns string, values []census2.Str) string {
	l.outfitIdsMu.Lock()
	defer l.outfitIdsMu.Unlock()
	l.outfitIdsOperand.Set(census2.NewList(values, ","))
	l.outfitIdsQuery.SetLimit(len(values))
	l.outfitIdsQuery.SetNamespace(ns)
	return l.client.ToURL(l.outfitIdsQuery)
}

func (l *Repository) OutfitIdsByTags(ctx context.Context, platform ps2_platforms.Platform, outfitTags []string) (map[string]ps2.OutfitId, error) {
	if len(outfitTags) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(outfitTags))
	for i, tag := range outfitTags {
		values[i] = census2.Str(strings.ToLower(tag))
	}
	url := l.outfitIdsUrl(ps2_platforms.PlatformNamespace(platform), values)
	outfits, err := census2_adapters.RetryableExecutePreparedAndDecode[ps2_collections.OutfitItem](
		ctx,
		l.log,
		l.client,
		ps2_collections.Outfit,
		url,
	)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]ps2.OutfitId, len(outfits))
	for _, outfit := range outfits {
		ids[outfit.Alias] = ps2.OutfitId(outfit.OutfitId)
	}
	return ids, nil
}
