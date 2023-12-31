package loaders

import (
	"context"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/voidwell"
)

type VoidWellWorldPopulationLoader struct {
	client *voidwell.Client
}

func NewVoidWellWorldPopulationLoader(client *voidwell.Client) *VoidWellWorldPopulationLoader {
	return &VoidWellWorldPopulationLoader{
		client: client,
	}
}

func (p *VoidWellWorldPopulationLoader) Name() string {
	return p.client.Endpoint()
}

func (p *VoidWellWorldPopulationLoader) Load(ctx context.Context, worldId ps2.WorldId) (ps2.WorldPopulation, error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return ps2.WorldPopulation{}, err
	}
	for _, state := range states {
		wId := ps2.WorldId(state.Id)
		if wId != worldId {
			continue
		}
		world := ps2.WorldPopulation{}
		world.Id = wId
		world.Name = state.Name
		zones := make(ps2.Zones, len(state.ZoneStates))
		world.Zones = zones
		for _, zoneState := range state.ZoneStates {
			zoneId := ps2.ZoneId(zoneState.Id)
			zone := ps2.ZonePopulation{
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
			zones[zoneId] = zone
			world.Total.All += zone.StatsByFactions.All
			world.Total.VS += zone.StatsByFactions.VS
			world.Total.NC += zone.StatsByFactions.NC
			world.Total.TR += zone.StatsByFactions.TR
			world.Total.NS += zone.StatsByFactions.NS
		}
		return world, nil
	}
	return ps2.WorldPopulation{}, ps2.ErrWorldNotFound
}
