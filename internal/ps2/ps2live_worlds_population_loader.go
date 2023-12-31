package ps2

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/ps2live"
)

type PS2LiveWorldsPopulationLoader struct {
	client *ps2live.PopulationClient
}

func NewPS2LiveWorldsPopulationLoader(client *ps2live.PopulationClient) *PS2LiveWorldsPopulationLoader {
	return &PS2LiveWorldsPopulationLoader{
		client: client,
	}
}

func (p *PS2LiveWorldsPopulationLoader) Load(ctx context.Context) (WorldsPopulation, error) {
	pops, err := p.client.AllPopulation(ctx)
	if err != nil {
		return WorldsPopulation{}, err
	}
	worlds := make(Worlds, len(pops))
	population := WorldsPopulation{
		Worlds: worlds,
	}
	for _, pop := range pops {
		worldId := WorldId(pop.Id)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = WorldNames[worldId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", worldId)
		}
		world.Total.All = pop.Average
		world.Total.VS = pop.Factions.VS
		world.Total.NC = pop.Factions.NC
		world.Total.TR = pop.Factions.TR
		worlds[worldId] = world
		population.Total.All += pop.Average
		population.Total.VS += pop.Factions.VS
		population.Total.NC += pop.Factions.NC
		population.Total.TR += pop.Factions.TR
	}
	return population, nil
}
