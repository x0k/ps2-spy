package sanctuary_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/sanctuary"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

var worldsPopulationQuery = sanctuary.NewQuery(sanctuary.GetQuery, sanctuary.Ns_ps2, sanctuary.WorldPopulationCollection)

func New(
	client *census2.Client,
) loader.Simple[meta.Loaded[ps2.WorldsPopulation]] {
	return func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
		wp, err := census2.ExecuteAndDecode[sanctuary.WorldPopulation](ctx, client, worldsPopulationQuery)
		if err != nil {
			return meta.Loaded[ps2.WorldsPopulation]{}, err
		}
		worlds := make([]ps2.WorldPopulation, len(wp))
		population := ps2.WorldsPopulation{
			Worlds: worlds,
		}
		for i, w := range wp {
			world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(w.WorldId)), "")
			world.All = w.Total
			world.VS = w.Population.VS
			world.NC = w.Population.NC
			world.TR = w.Population.TR
			world.NS = w.Population.NSO
			worlds[i] = world
			population.Total += world.All
		}
		return meta.LoadedNow(client.Endpoint(), population), nil
	}
}
