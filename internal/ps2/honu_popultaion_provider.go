package ps2

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-feed/internal/honu"
)

type HonuPopulationProvider struct {
	client *honu.Client
}

func NewHonuPopulationProvider(client *honu.Client) *HonuPopulationProvider {
	return &HonuPopulationProvider{
		client: client,
	}
}

func (p *HonuPopulationProvider) Name() string {
	return p.client.Endpoint()
}

func (p *HonuPopulationProvider) Population(ctx context.Context) (Population, error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return Population{}, err
	}
	worlds := make(Worlds, len(overview))
	population := Population{
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
				CommonPopulation: CommonPopulation{
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
