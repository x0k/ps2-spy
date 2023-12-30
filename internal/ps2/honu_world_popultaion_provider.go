package ps2

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/honu"
)

type HonuWorldPopulationProvider struct {
	client *honu.Client
}

func NewHonuWorldPopulationProvider(client *honu.Client) *HonuWorldPopulationProvider {
	return &HonuWorldPopulationProvider{
		client: client,
	}
}

func (p *HonuWorldPopulationProvider) Name() string {
	return p.client.Endpoint()
}

func (p *HonuWorldPopulationProvider) Load(ctx context.Context, worldId WorldId) (WorldPopulation, error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return WorldPopulation{}, err
	}
	for _, w := range overview {
		wId := WorldId(w.WorldId)
		if wId != worldId {
			continue
		}
		world := WorldPopulation{}
		world.Id = wId
		world.Name = WorldNames[wId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", wId)
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
		return world, nil
	}
	return WorldPopulation{}, ErrWorldNotFound
}
