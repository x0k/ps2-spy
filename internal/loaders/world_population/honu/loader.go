package honu_world_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	client *honu.Client,
) loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]] {
	return func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
		overview, err := client.WorldOverview(ctx)
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
			return meta.LoadedNow(client.Endpoint(), world), nil
		}
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
	}
}
