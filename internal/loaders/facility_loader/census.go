package facility_loader

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
		query: census2.NewQuery(census2.GetQuery, ns, collections.MapRegion).
			Where(census2.Cond("facility_id").Equals(operand)).
			Show("facility_id", "facility_name", "facility_type", "zone_id"),
	}
}

func (l *CensusLoader) toUrl(facilityId string) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(facilityId))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, facilityId string) (ps2.Facility, error) {
	url := l.toUrl(facilityId)
	outfits, err := census2.ExecutePreparedAndDecode[collections.MapRegionItem](
		ctx,
		l.client,
		collections.MapRegion,
		url,
	)
	if err != nil {
		return ps2.Facility{}, err
	}
	if len(outfits) == 0 {
		return ps2.Facility{}, loaders.ErrNotFound
	}
	zoneId, err := ps2.ToZoneId(outfits[0].ZoneId)
	if err != nil {
		return ps2.Facility{}, err
	}
	return ps2.Facility{
		Id:     outfits[0].FacilityId,
		Name:   outfits[0].FacilityName,
		Type:   outfits[0].FacilityType,
		ZoneId: zoneId,
	}, nil
}
