package census_outfits_repo

import (
	"context"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (l *Repository) outfitTagsUrl(ns string, values []census2.Str) string {
	l.outfitTagsMu.Lock()
	defer l.outfitTagsMu.Unlock()
	l.outfitTagsQuery.SetLimit(len(values))
	l.outfitTagsQuery.SetNamespace(ns)
	l.outfitTagsOperand.Set(census2.NewList(values, ","))
	return l.client.ToURL(l.outfitTagsQuery)
}

func (l *Repository) OutfitTagsByIds(
	ctx context.Context, platform ps2_platforms.Platform, outfitIds []ps2.OutfitId,
) (map[ps2.OutfitId]string, error) {
	if len(outfitIds) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(outfitIds))
	for i, outfitId := range outfitIds {
		values[i] = census2.Str(outfitId)
	}
	url := l.outfitTagsUrl(ps2_platforms.PlatformEnvironment(platform), values)
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
	tags := make(map[ps2.OutfitId]string, len(outfits))
	for _, outfit := range outfits {
		tags[ps2.OutfitId(outfit.OutfitId)] = outfit.Alias
	}
	return tags, nil
}
