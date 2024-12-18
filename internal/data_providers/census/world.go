package census_data_provider

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
)

func (l *DataProvider) worldMapUrl(ns string, worldId ps2.WorldId) string {
	l.worldMapMu.Lock()
	defer l.worldMapMu.Unlock()
	l.worldMapOperand.Set(census2.Str(worldId))
	l.worldMapQuery.SetNamespace(ns)
	return l.client.ToURL(l.worldMapQuery)
}

func (l *DataProvider) WorldMap(ctx context.Context, ns string, worldId ps2.WorldId) (ps2.WorldMap, error) {
	url := l.worldMapUrl(ns, worldId)
	zonesData, err := census2.ExecutePreparedAndDecode[ps2_collections.MapItem](
		ctx,
		l.client,
		ps2_collections.Map,
		url,
	)
	if err != nil {
		return ps2.WorldMap{}, fmt.Errorf("failed to get world map %q: %w", string(worldId), err)
	}
	zones := make(map[ps2.ZoneId]ps2.ZoneMap, len(zonesData))
	for _, zoneData := range zonesData {
		zoneId := ps2.ZoneId(zoneData.ZoneId)
		facilities := make(map[ps2.FacilityId]ps2_factions.Id, len(zoneData.Regions.Row))
		for _, region := range zoneData.Regions.Row {
			facilities[ps2.FacilityId(region.RowData.MapRegion.FacilityId)] =
				ps2_factions.Id(region.RowData.FactionId)
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
