package ps2

import (
	"context"

	"github.com/x0k/ps2-spy/internal/voidwell"
)

type VoidWellWorldsPopulationLoader struct {
	client *voidwell.Client
}

func NewVoidWellWorldsPopulationLoader(client *voidwell.Client) *VoidWellWorldsPopulationLoader {
	return &VoidWellWorldsPopulationLoader{
		client: client,
	}
}

func (p *VoidWellWorldsPopulationLoader) Load(ctx context.Context) (WorldsPopulation, error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return WorldsPopulation{}, err
	}
	worlds := make(Worlds, len(states))
	population := WorldsPopulation{
		Worlds: worlds,
	}
	for _, state := range states {
		worldId := WorldId(state.Id)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = state.Name
		zones := make(Zones, len(state.ZoneStates))
		world.Zones = zones
		for _, zoneState := range state.ZoneStates {
			zoneId := ZoneId(zoneState.Id)
			zone := ZonePopulation{
				Id:     zoneId,
				Name:   zoneState.Name,
				IsOpen: zoneState.LockState.State == "UNLOCKED",
				StatsByFactions: StatsByFactions{
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
