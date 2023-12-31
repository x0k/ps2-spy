package loaders

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2live/population"
)

type PS2LiveWorldsPopulationLoader struct {
	client *population.Client
}

func NewPS2LiveWorldsPopulationLoader(client *population.Client) *PS2LiveWorldsPopulationLoader {
	return &PS2LiveWorldsPopulationLoader{
		client: client,
	}
}

func (p *PS2LiveWorldsPopulationLoader) Name() string {
	return p.client.Endpoint()
}

func (p *PS2LiveWorldsPopulationLoader) Load(ctx context.Context) (ps2.WorldsPopulation, error) {
	pops, err := p.client.AllPopulation(ctx)
	if err != nil {
		return ps2.WorldsPopulation{}, err
	}
	worlds := make(ps2.Worlds, len(pops))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for _, pop := range pops {
		worldId := ps2.WorldId(pop.Id)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = ps2.WorldNames[worldId]
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
