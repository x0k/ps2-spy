package population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type SaerroLoader struct {
	client *saerro.Client
}

func NewSaerro(client *saerro.Client) *SaerroLoader {
	return &SaerroLoader{
		client: client,
	}
}

func (p *SaerroLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	data, err := p.client.AllWorldsPopulation(ctx)
	if err != nil {
		return loaders.Loaded[ps2.WorldsPopulation]{}, err
	}
	allWorlds := data.AllWorlds
	worlds := make([]ps2.WorldPopulation, len(allWorlds))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, w := range allWorlds {
		world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(w.Id)), w.Name)
		for _, z := range w.Zones.All {
			world.All += z.Population.Total
			world.VS += z.Population.VS
			world.NC += z.Population.NC
			world.TR += z.Population.TR
			world.NS += z.Population.NS
		}
		worlds[i] = world
		population.Total += world.All
	}
	return loaders.LoadedNow(p.client.Endpoint(), population), nil
}
