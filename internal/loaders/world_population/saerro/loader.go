package saerro_world_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	client *saerro.Client,
) loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]] {
	return func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
		data, err := client.AllWorldsPopulation(ctx)
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
			return meta.LoadedNow(client.Endpoint(), world), nil
		}
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
	}
}
