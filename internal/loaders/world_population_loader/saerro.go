package world_population_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type SaerroLoader struct {
	client *saerro.Client
}

func NewSaerro(client *saerro.Client) *SaerroLoader {
	return &SaerroLoader{
		client: client,
	}
}

func (p *SaerroLoader) Load(ctx context.Context, worldId ps2.WorldId) (loaders.Loaded[ps2.DetailedWorldPopulation], error) {
	data, err := p.client.AllWorldsPopulation(ctx)
	if err != nil {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, err
	}
	for _, w := range data.AllWorlds {
		wId := ps2.WorldId(w.Id)
		if wId != worldId {
			continue
		}
		world := ps2.DetailedWorldPopulation{
			Id:    wId,
			Name:  ps2.WorldNameById(wId),
			Zones: make([]ps2.ZonePopulation, len(w.Zones.All)),
		}
		for i, z := range w.Zones.All {
			zoneId := ps2.ZoneId(z.Id)
			world.Zones[i] = ps2.ZonePopulation{
				Id:   zoneId,
				Name: ps2.ZoneNameById(zoneId),
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
			world.Total += z.Population.Total
		}
		return loaders.LoadedNow(p.client.Endpoint(), world), nil
	}
	return loaders.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
}
