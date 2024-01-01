package loaders

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/sanctuary"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type SanctuaryWorldsPopulationLoader struct {
	client *census2.Client
}

func NewSanctuaryWorldsPopulationLoader(client *census2.Client) *SanctuaryWorldsPopulationLoader {
	return &SanctuaryWorldsPopulationLoader{
		client: client,
	}
}

func (l *SanctuaryWorldsPopulationLoader) Name() string {
	return l.client.Endpoint()
}

var WorldsPopulationQuery = sanctuary.NewQuery(sanctuary.GetQuery, sanctuary.Ns_ps2, sanctuary.WorldPopulationCollection)

func (l *SanctuaryWorldsPopulationLoader) Load(ctx context.Context) (ps2.WorldsPopulation, error) {
	wp, err := census2.ExecuteAndDecode[sanctuary.WorldPopulation](ctx, l.client, WorldsPopulationQuery)
	if err != nil {
		return ps2.WorldsPopulation{}, err
	}
	worlds := make(ps2.Worlds, len(wp))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for _, w := range wp {
		worldId := ps2.WorldId(w.WorldId)
		world := worlds[worldId]
		world.Id = worldId
		world.Name = ps2.WorldNames[worldId]
		if world.Name == "" {
			world.Name = fmt.Sprintf("World %d", worldId)
		}
		world.Total.All = w.Total
		world.Total.VS = w.Population.VS
		world.Total.NC = w.Population.NC
		world.Total.TR = w.Population.TR
		world.Total.NS = w.Population.NSO
		worlds[worldId] = world
		population.Total.All += world.Total.All
		population.Total.VS += w.Population.VS
		population.Total.NC += w.Population.NC
		population.Total.TR += w.Population.TR
		population.Total.NS += w.Population.NSO
	}
	return population, nil
}
