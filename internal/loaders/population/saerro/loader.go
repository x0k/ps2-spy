package saerro_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/ps2live/saerro"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(client *saerro.Client) loader.Simple[meta.Loaded[ps2.WorldsPopulation]] {
	return func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
		data, err := client.AllWorldsPopulation(ctx)
		if err != nil {
			return meta.Loaded[ps2.WorldsPopulation]{}, err
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
		return meta.LoadedNow(client.Endpoint(), population), nil
	}
}
