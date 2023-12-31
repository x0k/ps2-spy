package loaders

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type VoidWellWorldsPopulationLoader struct {
	client *voidwell.Client
}

func NewVoidWellWorldsPopulationLoader(client *voidwell.Client) *VoidWellWorldsPopulationLoader {
	return &VoidWellWorldsPopulationLoader{
		client: client,
	}
}

func (p *VoidWellWorldsPopulationLoader) Name() string {
	return p.client.Endpoint()
}

func (p *VoidWellWorldsPopulationLoader) Load(ctx context.Context) (ps2.WorldsPopulation, error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return ps2.WorldsPopulation{}, err
	}
	worlds := make(ps2.Worlds, len(states))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for _, state := range states {
		worldId := ps2.WorldId(state.Id)
		world := worlds[worldId]
		world.Id = worldId
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
		worlds[worldId] = world
		population.Total.All += world.Total.All
		population.Total.VS += world.Total.VS
		population.Total.NC += world.Total.NC
		population.Total.TR += world.Total.TR
		population.Total.NS += world.Total.NS
	}
	return population, nil
}
