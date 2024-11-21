package voidwell_world_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	client *voidwell.Client,
) loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]] {
	return func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
		states, err := client.WorldsState(ctx)
		if err != nil {
			return meta.Loaded[ps2.DetailedWorldPopulation]{}, err
		}
		for _, state := range states {
			wId := ps2.WorldId(strconv.Itoa(state.Id))
			if wId != worldId {
				continue
			}
			world := ps2.DetailedWorldPopulation{
				Id:    wId,
				Name:  ps2.WorldNameById(wId),
				Zones: make([]ps2.ZonePopulation, len(state.ZoneStates)),
			}
			for i, zoneState := range state.ZoneStates {
				zoneId := ps2.ZoneId(strconv.Itoa(zoneState.Id))
				world.Zones[i] = ps2.ZonePopulation{
					Id:     zoneId,
					Name:   zoneState.Name,
					IsOpen: zoneState.LockState.State == "UNLOCKED",
					StatPerFactions: ps2.StatPerFactions{
						All: zoneState.Population.NC + zoneState.Population.TR + zoneState.Population.VS + zoneState.Population.NS,
						VS:  zoneState.Population.VS,
						NC:  zoneState.Population.NC,
						TR:  zoneState.Population.TR,
						NS:  zoneState.Population.NS,
					},
				}
				world.Total += world.Zones[i].StatPerFactions.All
			}
			return meta.LoadedNow(client.Endpoint(), world), nil
		}
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
	}
}
