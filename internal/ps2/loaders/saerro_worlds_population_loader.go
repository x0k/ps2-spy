package loaders

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type SaerroWorldsPopulationLoader struct {
	client *saerro.Client
}

func NewSaerroWorldsPopulationLoader(client *saerro.Client) *SaerroWorldsPopulationLoader {
	return &SaerroWorldsPopulationLoader{
		client: client,
	}
}

func (p *SaerroWorldsPopulationLoader) Name() string {
	return p.client.Endpoint()
}

func (p *SaerroWorldsPopulationLoader) Load(ctx context.Context) (ps2.WorldsPopulation, error) {
	data, err := p.client.AllWorldsPopulation(ctx)
	if err != nil {
		return ps2.WorldsPopulation{}, err
	}
	allWorlds := data.AllWorlds
	worlds := make(ps2.Worlds, len(allWorlds))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for _, w := range allWorlds {
		worldId := ps2.WorldId(w.Id)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = ps2.WorldNames[worldId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", worldId)
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
		worlds[worldId] = world
		population.Total.All += world.Total.All
		population.Total.VS += world.Total.VS
		population.Total.NC += world.Total.NC
		population.Total.TR += world.Total.TR
		population.Total.NS += world.Total.NS
	}
	return population, nil
}
