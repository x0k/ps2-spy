package world_population_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type VoidWellLoader struct {
	client *voidwell.Client
}

func NewVoidWell(client *voidwell.Client) *VoidWellLoader {
	return &VoidWellLoader{
		client: client,
	}
}

func (p *VoidWellLoader) Load(ctx context.Context, worldId ps2.WorldId) (loaders.Loaded[ps2.DetailedWorldPopulation], error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return loaders.Loaded[ps2.DetailedWorldPopulation]{}, err
	}
	for _, state := range states {
		wId := ps2.WorldId(state.Id)
		if wId != worldId {
			continue
		}
		world := ps2.DetailedWorldPopulation{
			Id:    wId,
			Name:  ps2.WorldNameById(wId),
			Zones: make([]ps2.ZonePopulation, len(state.ZoneStates)),
		}
		for i, zoneState := range state.ZoneStates {
			zoneId := ps2.ZoneId(zoneState.Id)
			world.Zones[i] = ps2.ZonePopulation{
				Id:     zoneId,
				Name:   zoneState.Name,
				IsOpen: zoneState.LockState.State == "UNLOCKED",
				StatsByFactions: ps2.StatsByFactions{
					All: zoneState.Population.NC + zoneState.Population.TR + zoneState.Population.VS + zoneState.Population.NS,
					VS:  zoneState.Population.VS,
					NC:  zoneState.Population.NC,
					TR:  zoneState.Population.TR,
					NS:  zoneState.Population.NS,
				},
			}
			world.Total += world.Zones[i].StatsByFactions.All
		}
		return loaders.LoadedNow(p.client.Endpoint(), world), nil
	}
	return loaders.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
}
