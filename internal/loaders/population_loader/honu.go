package population_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type HonuLoader struct {
	client *honu.Client
}

func NewHonu(client *honu.Client) *HonuLoader {
	return &HonuLoader{
		client: client,
	}
}

func (p *HonuLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return loaders.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, len(overview))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, w := range overview {
		world := ps2.NewWorldPopulation(ps2.WorldId(w.WorldId), w.WorldName)
		for _, z := range w.Zones {
			world.All += z.Players.All
			world.VS += z.Players.VS
			world.NC += z.Players.NC
			world.TR += z.Players.TR
			world.Other += z.Players.Unknown
		}
		worlds[i] = world
		population.Total += world.All
	}
	return loaders.LoadedNow(p.client.Endpoint(), population), nil
}
