package voidwell_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/voidwell"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	client *voidwell.Client,
) loader.Simple[meta.Loaded[ps2.WorldsPopulation]] {
	return func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
		states, err := client.WorldsState(ctx)
		if err != nil {
			return meta.Loaded[ps2.WorldsPopulation]{}, err
		}
		worlds := make([]ps2.WorldPopulation, len(states))
		population := ps2.WorldsPopulation{
			Worlds: worlds,
		}
		for i, state := range states {
			world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(state.Id)), state.Name)
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
		return meta.LoadedNow(client.Endpoint(), population), nil
	}
}
