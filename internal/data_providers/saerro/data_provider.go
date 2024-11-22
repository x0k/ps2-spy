package saerro_data_provider

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type DataProvider struct {
	client *saerro.Client
}

func New(client *saerro.Client) *DataProvider {
	return &DataProvider{
		client: client,
	}
}

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	data, err := p.client.AllWorldsPopulation(ctx)
	if err != nil {
		return meta.Loaded[ps2.WorldsPopulation]{}, err
	}
	allWorlds := data.AllWorlds
	worlds := make([]ps2.WorldPopulation, len(allWorlds))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, w := range allWorlds {
		world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(w.Id)), w.Name)
		for _, z := range w.Zones.All {
			world.All += z.Population.Total
			world.VS += z.Population.VS
			world.NC += z.Population.NC
			world.TR += z.Population.TR
			world.NS += z.Population.NS
		}
		worlds[i] = world
		population.Total += world.All
	}
	return meta.LoadedNow(p.client.Endpoint(), population), nil
}

func (p *DataProvider) WorldPopulation(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
	data, err := p.client.AllWorldsPopulation(ctx)
	if err != nil {
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, err
	}
	for _, w := range data.AllWorlds {
		wId := ps2.WorldId(strconv.Itoa(w.Id))
		if wId != worldId {
			continue
		}
		world := ps2.DetailedWorldPopulation{
			Id:    wId,
			Name:  ps2.WorldNameById(wId),
			Zones: make([]ps2.ZonePopulation, len(w.Zones.All)),
		}
		for i, z := range w.Zones.All {
			zoneId := ps2.ZoneId(strconv.Itoa(z.Id))
			world.Zones[i] = ps2.ZonePopulation{
				Id:   zoneId,
				Name: ps2.ZoneNameById(zoneId),
				// TODO: fix this by loading additional data
				IsOpen: z.Population.Total != 0,
				StatPerFactions: ps2.StatPerFactions{
					All: z.Population.Total,
					VS:  z.Population.VS,
					NC:  z.Population.NC,
					TR:  z.Population.TR,
					NS:  z.Population.NS,
				},
			}
			world.Total += z.Population.Total
		}
		return meta.LoadedNow(p.client.Endpoint(), world), nil
	}
	return meta.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
}
