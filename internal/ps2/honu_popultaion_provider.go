package ps2

import (
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

func (p *HonuPopulationProvider) Population() (Population, error) {
	worlds, err := p.client.WorldOverview()
	if err != nil {
		return nil, err
	}
	population := make(Population, len(worlds))
	for _, w := range worlds {
		worldId := WorldId(w.WorldId)
		zones := make(Zones, len(w.Zones))
		world := WorldPopulation{
			WorldId: worldId,
			Zones:   zones,
		}
		for _, z := range w.Zones {
			zoneId := ZoneId(z.ZoneId)
			zones[zoneId] = ZonePopulation{
				ZoneId: zoneId,
				IsOpen: z.IsOpened,
				All:    z.Players.All,
				VS:     z.Players.VS,
				NC:     z.Players.NC,
				TR:     z.Players.TR,
				Other:  z.Players.Unknown,
			}
			world.Total.All += z.Players.All
			world.Total.VS += z.Players.VS
			world.Total.NC += z.Players.NC
			world.Total.TR += z.Players.TR
			world.Total.Other += z.Players.Unknown
		}
	}
	return population, nil
}
