package loaders

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type SaerroWorldPopulationLoader struct {
	client *saerro.Client
}

func NewSaerroWorldPopulationLoader(client *saerro.Client) *SaerroWorldPopulationLoader {
	return &SaerroWorldPopulationLoader{
		client: client,
	}
}

func (p *SaerroWorldPopulationLoader) Name() string {
	return p.client.Endpoint()
}

func (p *SaerroWorldPopulationLoader) Load(ctx context.Context, worldId ps2.WorldId) (ps2.WorldPopulation, error) {
	data, err := p.client.AllWorldsPopulation(ctx)
	if err != nil {
		return ps2.WorldPopulation{}, err
	}
	allWorlds := data.AllWorlds
	for _, w := range allWorlds {
		wId := ps2.WorldId(w.Id)
		if wId != worldId {
			continue
		}
		world := ps2.WorldPopulation{}
		world.Id = wId
		world.Name = ps2.WorldNames[wId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", wId)
		}
		allZones := w.Zones.All
		zones := make(ps2.Zones, len(allZones))
		world.Zones = zones
		for _, z := range allZones {
			zoneId := ps2.ZoneId(z.Id)
			zone := ps2.ZonePopulation{
				Id:   zoneId,
				Name: ps2.ZoneNames[zoneId],
				// TODO: fix this by loading additional data
				IsOpen: z.Population.Total != 0,
				StatsByFactions: ps2.StatsByFactions{
					All: z.Population.Total,
					VS:  z.Population.VS,
					NC:  z.Population.NC,
					TR:  z.Population.TR,
					NS:  z.Population.NS,
				},
			}
			if zone.Name == "" {
				zone.Name = fmt.Sprintf("Zone %d", zoneId)
			}
			world.Zones[zoneId] = zone
			world.Total.All += z.Population.Total
			world.Total.VS += z.Population.VS
			world.Total.NC += z.Population.NC
			world.Total.TR += z.Population.TR
			world.Total.NS += z.Population.NS
		}
		return world, nil
	}
	return ps2.WorldPopulation{}, ps2.ErrWorldNotFound
}
