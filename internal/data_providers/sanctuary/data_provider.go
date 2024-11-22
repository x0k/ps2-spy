package sanctuary_data_provider

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/sanctuary"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type DataProvider struct {
	client *census2.Client
}

func New(client *census2.Client) *DataProvider {
	return &DataProvider{
		client: client,
	}
}

var worldsPopulationQuery = sanctuary.NewQuery(sanctuary.GetQuery, sanctuary.Ns_ps2, sanctuary.WorldPopulationCollection)

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	wp, err := census2.ExecuteAndDecode[sanctuary.WorldPopulation](ctx, p.client, worldsPopulationQuery)
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
	return meta.LoadedNow(p.client.Endpoint(), population), nil
}
