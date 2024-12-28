package census_data_provider

import (
	"context"

	census2_adapters "github.com/x0k/ps2-spy/internal/adapters/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/shared"
)

func (l *DataProvider) facilityUrl(ns string, facilityId ps2.FacilityId) string {
	l.facilityMu.Lock()
	defer l.facilityMu.Unlock()
	l.facilityOperand.Set(census2.Str(facilityId))
	l.facilityQuery.SetNamespace(ns)
	return l.client.ToURL(l.facilityQuery)
}

func (l *DataProvider) Facility(ctx context.Context, ns string, facilityId ps2.FacilityId) (ps2.Facility, error) {
	url := l.facilityUrl(ns, facilityId)
	regions, err := census2_adapters.RetryableExecutePreparedAndDecode[ps2_collections.MapRegionItem](
		ctx,
		l.log,
		l.client,
		ps2_collections.MapRegion,
		url,
	)
	if err != nil {
		return ps2.Facility{}, err
	}
	if len(regions) == 0 {
		return ps2.Facility{}, shared.ErrNotFound
	}
	return ps2.Facility{
		Id:     ps2.FacilityId(regions[0].FacilityId),
		Name:   regions[0].FacilityName,
		Type:   regions[0].FacilityType,
		ZoneId: ps2.ZoneId(regions[0].ZoneId),
	}, nil
}
