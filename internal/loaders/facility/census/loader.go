package census_facility_loader

import (
	"context"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/shared"
)

type Loader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand *census2.Ptr[census2.Str]
}

func New(client *census2.Client, ns string) *Loader {
	operand := census2.NewPtr(census2.Str(""))
	return &Loader{
		client:  client,
		operand: &operand,
		query: census2.NewQuery(census2.GetQuery, ns, ps2_collections.MapRegion).
			Where(census2.Cond("facility_id").Equals(&operand)).
			Show("facility_id", "facility_name", "facility_type", "zone_id"),
	}
}

func (l *Loader) toUrl(facilityId ps2.FacilityId) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(facilityId))
	return l.client.ToURL(l.query)
}

func (l *Loader) Load(ctx context.Context, facilityId ps2.FacilityId) (ps2.Facility, error) {
	url := l.toUrl(facilityId)
	regions, err := census2.ExecutePreparedAndDecode[ps2_collections.MapRegionItem](
		ctx,
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
