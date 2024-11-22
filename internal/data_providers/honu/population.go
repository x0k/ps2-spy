package honu_data_provider

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return meta.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, len(overview))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, w := range overview {
		world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(w.WorldId)), w.WorldName)
		for _, z := range w.Zones {
			world.All += z.Players.All
			world.VS += z.Players.VS
			world.NC += z.Players.NC
			world.TR += z.Players.TR
			world.Other += z.Players.Unknown
		}
		worlds[i] = world
		population.Total += world.All
	}
	return meta.LoadedNow(p.client.Endpoint(), population), nil
}

func (p *DataProvider) WorldPopulation(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, err
	}
	for _, w := range overview {
		wId := ps2.WorldId(strconv.Itoa(w.WorldId))
		if wId != worldId {
			continue
		}
		world := ps2.DetailedWorldPopulation{
			Id:    wId,
			Name:  ps2.WorldNameById(wId),
			Zones: make([]ps2.ZonePopulation, len(w.Zones)),
		}
		for i, z := range w.Zones {
			zoneId := ps2.ZoneId(strconv.Itoa(z.ZoneId))
			world.Zones[i] = ps2.ZonePopulation{
				Id:     zoneId,
				Name:   ps2.ZoneNameById(zoneId),
				IsOpen: z.IsOpened,
				StatPerFactions: ps2.StatPerFactions{
					All:   z.Players.All,
					VS:    z.Players.VS,
					NC:    z.Players.NC,
					TR:    z.Players.TR,
					Other: z.Players.Unknown,
				},
			}
			world.Total += z.Players.All
		}
		return meta.LoadedNow(p.client.Endpoint(), world), nil
	}
	return meta.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
}
