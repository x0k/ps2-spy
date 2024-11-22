package fisu_data_provider

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/fisu"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type DataProvider struct {
	client *fisu.Client
}

func New(
	client *fisu.Client,
) *DataProvider {
	return &DataProvider{
		client: client,
	}
}

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	worldsPopulation, err := p.client.WorldsPopulation(ctx)
	if err != nil {
		return meta.Loaded[ps2.WorldsPopulation]{}, err
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
	return meta.LoadedNow(p.client.Endpoint(), population), nil
}
