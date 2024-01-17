package worlds_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type VoidWellLoader struct {
	client *voidwell.Client
}

func NewVoidWell(client *voidwell.Client) *VoidWellLoader {
	return &VoidWellLoader{
		client: client,
	}
}

func (p *VoidWellLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	states, err := p.client.WorldsState(ctx)
	if err != nil {
		return loaders.Loaded[ps2.WorldsPopulation]{}, err
	}
	worlds := make([]ps2.WorldPopulation, len(states))
	population := ps2.WorldsPopulation{
		Worlds: worlds,
	}
	for i, state := range states {
		world := ps2.NewWorldPopulation(ps2.WorldId(state.Id), state.Name)
		for _, zoneState := range state.ZoneStates {
			world.All += zoneState.Population.NC +
				zoneState.Population.TR +
				zoneState.Population.VS +
				zoneState.Population.NS
			world.VS += zoneState.Population.VS
			world.NC += zoneState.Population.NC
			world.TR += zoneState.Population.TR
			world.NS += zoneState.Population.NS
		}
		worlds[i] = world
		population.Total += world.All
	}
	return loaders.LoadedNow(p.client.Endpoint(), population), nil
}
