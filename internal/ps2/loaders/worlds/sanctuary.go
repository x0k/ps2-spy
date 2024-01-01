package worlds

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/sanctuary"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type SanctuaryLoader struct {
	client *census2.Client
}

func NewSanctuaryLoader(client *census2.Client) *SanctuaryLoader {
	return &SanctuaryLoader{
		client: client,
	}
}

var WorldsPopulationQuery = sanctuary.NewQuery(sanctuary.GetQuery, sanctuary.Ns_ps2, sanctuary.WorldPopulationCollection)

func (l *SanctuaryLoader) Load(ctx context.Context) (ps2.Loaded[ps2.WorldsPopulation], error) {
	wp, err := census2.ExecuteAndDecode[sanctuary.WorldPopulation](ctx, l.client, WorldsPopulationQuery)
	if err != nil {
		return ps2.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, len(wp))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, w := range wp {
		world := ps2.NewWorldPopulation(ps2.WorldId(w.WorldId), "")
		world.All = w.Total
		world.VS = w.Population.VS
		world.NC = w.Population.NC
		world.TR = w.Population.TR
		world.NS = w.Population.NSO
		worlds[i] = world
		population.Total += world.All
	}
	return ps2.LoadedNow(l.client.Endpoint(), population), nil
}
