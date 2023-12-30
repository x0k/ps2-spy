package ps2

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/honu"
)

type HonuWorldsPopulationProvider struct {
	client *honu.Client
}

func NewHonuWorldsPopulationProvider(client *honu.Client) *HonuWorldsPopulationProvider {
	return &HonuWorldsPopulationProvider{
		client: client,
	}
}

func (p *HonuWorldsPopulationProvider) Name() string {
	return p.client.Endpoint()
}

func (p *HonuWorldsPopulationProvider) Load(ctx context.Context) (WorldsPopulation, error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return WorldsPopulation{}, err
	}
	worlds := make(Worlds, len(overview))
	population := WorldsPopulation{
		Worlds: worlds,
	}
	for _, w := range overview {
		worldId := WorldId(w.WorldId)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = WorldNames[worldId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", worldId)
		}
		zones := make(Zones, len(w.Zones))
		world.Zones = zones
		for _, z := range w.Zones {
			zoneId := ZoneId(z.ZoneId)
			zone := ZonePopulation{
				Id:     zoneId,
				Name:   ZoneNames[zoneId],
				IsOpen: z.IsOpened,
				StatsByFactions: StatsByFactions{
					All:   z.Players.All,
					VS:    z.Players.VS,
					NC:    z.Players.NC,
					TR:    z.Players.TR,
					Other: z.Players.Unknown,
				},
			}
			if zone.Name == "" {
				zone.Name = fmt.Sprintf("Zone %d", zoneId)
			}
			zones[zoneId] = zone
			world.Total.All += z.Players.All
			world.Total.VS += z.Players.VS
			world.Total.NC += z.Players.NC
			world.Total.TR += z.Players.TR
			world.Total.Other += z.Players.Unknown
		}
		worlds[worldId] = world
		population.Total.All += world.Total.All
		population.Total.VS += world.Total.VS
		population.Total.NC += world.Total.NC
		population.Total.TR += world.Total.TR
		population.Total.Other += world.Total.Other
	}
	return population, nil
}