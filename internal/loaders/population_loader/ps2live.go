package population_loader

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/ps2live/population"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type PS2LiveLoader struct {
	client *population.Client
}

func NewPS2Live(client *population.Client) *PS2LiveLoader {
	return &PS2LiveLoader{
		client: client,
	}
}

func (p *PS2LiveLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	pops, err := p.client.AllPopulation(ctx)
	if err != nil {
		return loaders.Loaded[ps2.WorldsPopulation]{}, err
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
		world := ps2.NewWorldPopulation(ps2.WorldId(pop.Id), "")
		world.All = pop.Average
		world.VS = pop.Factions.VS
		world.NC = pop.Factions.NC
		world.TR = pop.Factions.TR
		worlds[i] = world
		population.Total += pop.Average
	}
	return loaders.Loaded[ps2.WorldsPopulation]{
		Value:     population,
		Source:    p.client.Endpoint(),
		UpdatedAt: updatedAt,
	}, nil
}
