package ps2spy_data_provider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (p *DataProvider) Population(ctx context.Context) (meta.Loaded[ps2.WorldsPopulation], error) {
	total := 0
	worlds := make([]ps2.WorldPopulation, 0)
	for _, platform := range ps2_platforms.Platforms {
		population, err := p.charactersTracker.WorldsPopulation(platform)
		if err != nil {
			p.log.Error(ctx, "failed to retrieve worlds population", slog.String("platform", string(platform)), sl.Err(err))
			continue
		}
		total += population.Total
		worlds = append(worlds, population.Worlds...)
	}
	return meta.LoadedNow(p.appName, ps2.WorldsPopulation{
		Total:  total,
		Worlds: worlds,
	}), nil
}

func (p *DataProvider) WorldPopulation(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
	platform, ok := ps2.WorldPlatforms[worldId]
	if !ok {
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("unknown world %q", worldId)
	}
	population, err := p.charactersTracker.DetailedWorldPopulation(platform, worldId)
	if err != nil {
		return meta.Loaded[ps2.DetailedWorldPopulation]{}, fmt.Errorf("getting population: %w", err)
	}
	return meta.LoadedNow(p.appName, population), nil
}
