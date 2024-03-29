package world_map_loader

import (
	"context"
	"strings"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	client  *census2.Client
	queryMu sync.Mutex
	query   *census2.Query
	operand census2.Ptr[census2.Str]
}

func NewCensus(client *census2.Client, platform platforms.Platform) *CensusLoader {
	operand := census2.NewPtr(census2.Str(""))
	ns := platforms.PlatformNamespace(platform)
	b := strings.Builder{}
	b.Grow(len(ps2.ZoneIds) * 3)
	b.WriteString(string(ps2.ZoneIds[0]))
	for _, zoneId := range ps2.ZoneIds[1:] {
		b.WriteString(",")
		b.WriteString(string(zoneId))
	}
	return &CensusLoader{
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, ns, collections.Map).
			Where(
				census2.Cond("world_id").Equals(operand),
				census2.Cond("zone_ids").Equals(census2.Str(b.String())),
			).
			WithJoin(
				census2.Join(collections.MapRegion).
					Show("facility_id").
					InjectAt("map_region").
					On("Regions.Row.RowData.RegionId").
					To("map_region_id"),
			),
	}
}

func (l *CensusLoader) toUrl(worldId ps2.WorldId) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.Str(worldId))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, worldId ps2.WorldId) (ps2.WorldMap, error) {
	url := l.toUrl(worldId)
	zonesData, err := census2.ExecutePreparedAndDecode[collections.MapItem](
		ctx,
		l.client,
		collections.Map,
		url,
	)
	if err != nil {
		return ps2.WorldMap{}, err
	}
	zones := make(map[ps2.ZoneId]ps2.ZoneMap, len(zonesData))
	for _, zoneData := range zonesData {
		zoneId := ps2.ZoneId(zoneData.ZoneId)
		facilities := make(map[ps2.FacilityId]factions.Id, len(zoneData.Regions.Row))
		for _, region := range zoneData.Regions.Row {
			facilities[ps2.FacilityId(region.RowData.MapRegion.FacilityId)] =
				factions.Id(region.RowData.FactionId)
		}
		zones[zoneId] = ps2.ZoneMap{
			Id:         zoneId,
			Facilities: facilities,
		}
	}
	return ps2.WorldMap{
		Id:    worldId,
		Zones: zones,
	}, nil
}
