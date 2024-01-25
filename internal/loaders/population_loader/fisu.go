package population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type FisuLoader struct {
	client *fisu.Client
}

func NewFisu(client *fisu.Client) *FisuLoader {
	return &FisuLoader{
		client: client,
	}
}

func (l *FisuLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	worldsPopulation, err := l.client.WorldsPopulation(ctx)
	if err != nil {
		return loaders.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, 0, len(worldsPopulation))
	population := ps2.WorldsPopulation{}
	for _, wpArr := range worldsPopulation {
		if len(wpArr) == 0 {
			continue
		}
		wp := wpArr[0]
		worldId := ps2.WorldId(strconv.Itoa(wp.WorldId))
		world := ps2.NewWorldPopulation(worldId, "")
		world.All = wp.VS + wp.NC + wp.TR + wp.NS
		world.VS = wp.VS
		world.NC = wp.NC
		world.TR = wp.TR
		world.NS = wp.NS
		world.Other = wp.Unknown
		worlds = append(worlds, world)
		population.Total += world.All
	}
	population.Worlds = worlds
	return loaders.LoadedNow(l.client.Endpoint(), population), nil
}
