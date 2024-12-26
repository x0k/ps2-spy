package census_data_provider

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (l *DataProvider) outfitsUrl(ns string, values []census2.Str) string {
	l.outfitsMu.Lock()
	defer l.outfitsMu.Unlock()
	l.outfitsQuery.SetLimit(len(values))
	l.outfitsQuery.SetNamespace(ns)
	l.outfitsOperand.Set(census2.NewList(values, ","))
	return l.client.ToURL(l.outfitsQuery)
}

func (l *DataProvider) Outfits(ctx context.Context, platform ps2_platforms.Platform, outfitIds []ps2.OutfitId) (map[ps2.OutfitId]ps2.Outfit, error) {
	if len(outfitIds) == 0 {
		return nil, nil
	}
	values := make([]census2.Str, len(outfitIds))
	for i, outfitId := range outfitIds {
		values[i] = census2.Str(outfitId)
	}
	url := l.outfitsUrl(ps2_platforms.PlatformNamespace(platform), values)
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
			Platform: platform,
		}
	}
	return res, nil
}
