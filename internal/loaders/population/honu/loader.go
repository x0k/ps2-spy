package honu_population_loader

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(client *honu.Client) loader.Simple[meta.Loaded[ps2.WorldsPopulation]] {
	return func(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
		overview, err := client.WorldOverview(ctx)
		if err != nil {
			return meta.Loaded[ps2.WorldsPopulation]{}, err
		}
		worlds := make([]ps2.WorldPopulation, len(overview))
		population := ps2.WorldsPopulation{
			Worlds: worlds,
		}
		for i, w := range overview {
			world := ps2.NewWorldPopulation(ps2.WorldId(strconv.Itoa(w.WorldId)), w.WorldName)
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
		return meta.LoadedNow(client.Endpoint(), population), nil
	}
}
