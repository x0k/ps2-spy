package loaders

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type FisuWorldsPopulationLoader struct {
	client *fisu.Client
}

func NewFisuWorldsPopulationLoader(client *fisu.Client) *FisuWorldsPopulationLoader {
	return &FisuWorldsPopulationLoader{
		client: client,
	}
}

func (l *FisuWorldsPopulationLoader) Name() string {
	return l.client.Endpoint()
}

func (l *FisuWorldsPopulationLoader) Load(ctx context.Context) (ps2.WorldsPopulation, error) {
	worldsPopulation, err := l.client.WorldsPopulation(ctx)
	if err != nil {
		return ps2.WorldsPopulation{}, err
	}
	worlds := make(ps2.Worlds, len(worldsPopulation))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for _, wArr := range worldsPopulation {
		if len(wArr) == 0 {
			continue
		}
		w := wArr[0]
		worldId := ps2.WorldId(w.WorldId)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = ps2.WorldNames[worldId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", worldId)
		}
		world.Total.All = w.VS + w.NC + w.TR + w.NS
		world.Total.VS = w.VS
		world.Total.NC = w.NC
		world.Total.TR = w.TR
		world.Total.NS = w.NS
		world.Total.Other = w.Unknown
		worlds[worldId] = world
		population.Total.All += world.Total.All
		population.Total.VS += w.VS
		population.Total.NC += w.NC
		population.Total.TR += w.TR
		population.Total.NS += w.NS
		population.Total.Other += w.Unknown
	}
	return population, nil
}
