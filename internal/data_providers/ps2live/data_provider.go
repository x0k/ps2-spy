package ps2live_data_provider

import (
	"context"
	"strconv"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type DataProvider struct {
	client *population.Client
}

func New(client *population.Client) *DataProvider {
	return &DataProvider{
		client: client,
	}
}

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	pops, err := p.client.AllPopulation(ctx)
	if err != nil {
		return meta.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, len(pops))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	updatedAt := time.Now()
	for i, pop := range pops {
		if cachedAt, err := time.Parse(time.RFC3339, pop.CachedAt); err == nil && cachedAt.Before(updatedAt) {
			updatedAt = cachedAt
		}
		world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(pop.Id)), "")
		world.All = pop.Average
		world.VS = pop.Factions.VS
		world.NC = pop.Factions.NC
		world.TR = pop.Factions.TR
		worlds[i] = world
		population.Total += pop.Average
	}
	return meta.Loaded[ps2.WorldsPopulation]{
		Value:     population,
		Source:    p.client.Endpoint(),
		UpdatedAt: updatedAt,
	}, nil
}
