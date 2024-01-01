package world

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/honu"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type HonuLoader struct {
	client *honu.Client
}

func NewHonuLoader(client *honu.Client) *HonuLoader {
	return &HonuLoader{
		client: client,
	}
}

func (p *HonuLoader) Name() string {
	return p.client.Endpoint()
}

func (p *HonuLoader) Load(ctx context.Context, worldId ps2.WorldId) (ps2.Loaded[ps2.DetailedWorldPopulation], error) {
	overview, err := p.client.WorldOverview(ctx)
	if err != nil {
		return ps2.Loaded[ps2.DetailedWorldPopulation]{}, err
	}
	for _, w := range overview {
		wId := ps2.WorldId(w.WorldId)
		if wId != worldId {
			continue
		}
		world := ps2.DetailedWorldPopulation{
			Id:    wId,
			Name:  ps2.WorldNameById(wId),
			Zones: make([]ps2.ZonePopulation, len(w.Zones)),
		}
		for i, z := range w.Zones {
			zoneId := ps2.ZoneId(z.ZoneId)
			world.Zones[i] = ps2.ZonePopulation{
				Id:     zoneId,
				Name:   ps2.ZoneNameById(zoneId),
				IsOpen: z.IsOpened,
				StatsByFactions: ps2.StatsByFactions{
					All:   z.Players.All,
					VS:    z.Players.VS,
					NC:    z.Players.NC,
					TR:    z.Players.TR,
					Other: z.Players.Unknown,
				},
			}
			world.Total += z.Players.All
		}
		return ps2.LoadedNow(p.client.Endpoint(), world), nil
	}
	return ps2.Loaded[ps2.DetailedWorldPopulation]{}, ps2.ErrWorldNotFound
}
