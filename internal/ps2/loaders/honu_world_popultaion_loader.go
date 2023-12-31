package loaders

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/honu"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type HonuWorldPopulationLoader struct {
	client *honu.Client
}

func NewHonuWorldPopulationLoader(client *honu.Client) *HonuWorldPopulationLoader {
	return &HonuWorldPopulationLoader{
		client: client,
	}
}

func (p *HonuWorldPopulationLoader) Name() string {
	return p.client.Endpoint()
}

func (p *HonuWorldPopulationLoader) Load(ctx context.Context, worldId ps2.WorldId) (ps2.WorldPopulation, error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return ps2.WorldPopulation{}, err
	}
	for _, w := range overview {
		wId := ps2.WorldId(w.WorldId)
		if wId != worldId {
			continue
		}
		world := ps2.WorldPopulation{}
		world.Id = wId
		world.Name = ps2.WorldNames[wId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", wId)
		}
		zones := make(ps2.Zones, len(w.Zones))
		world.Zones = zones
		for _, z := range w.Zones {
			zoneId := ps2.ZoneId(z.ZoneId)
			zone := ps2.ZonePopulation{
				Id:     zoneId,
				Name:   ps2.ZoneNames[zoneId],
				IsOpen: z.IsOpened,
				StatsByFactions: ps2.StatsByFactions{
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
	return ps2.WorldPopulation{}, ps2.ErrWorldNotFound
}
