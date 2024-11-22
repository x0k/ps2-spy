package voidwell_data_provider

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return meta.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, len(states))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, state := range states {
		world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(state.Id)), state.Name)
		for _, zoneState := range state.ZoneStates {
			world.All += zoneState.Population.NC +
				zoneState.Population.TR +
				zoneState.Population.VS +
				zoneState.Population.NS
			world.VS += zoneState.Population.VS
			world.NC += zoneState.Population.NC
			world.TR += zoneState.Population.TR
			world.NS += zoneState.Population.NS
		}
		worlds[i] = world
		population.Total += world.All
	}
	return meta.LoadedNow(p.client.Endpoint(), population), nil
}

func (p *DataProvider) WorldPopulation(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
	states, err := p.client.WorldsState(ctx)
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
		return meta.LoadedNow(p.client.Endpoint(), world), nil
	}
	return meta.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
}
