package discord_commands

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/cache/memory"
	"github.com/x0k/ps2-spy/internal/lib/containers"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type populationLoader struct {
	fallbacks *containers.Fallbacks[loader.Simple[meta.Loaded[ps2.WorldsPopulation]]]
	load      loader.Keyed[string, meta.Loaded[ps2.WorldsPopulation]]
}

func newPopulationLoader(
	log *logger.Logger,
	loaders map[string]loader.Simple[meta.Loaded[ps2.WorldsPopulation]],
	loadersPriority []string,
) *populationLoader {
	fallbacks := containers.NewFallbacks(
		log.Logger.With(sl.Component("population_loader_fallbacks")),
		loaders,
		loadersPriority,
		time.Hour,
	)
	fallbackLoader := loader.NewFallback(fallbacks)
	cached := loader.WithQueriedCache(
		log.Logger.With(sl.Component("population_loader_cache")),
		func(ctx context.Context, provider string) (meta.Loaded[ps2.WorldsPopulation], error) {
			if loader, ok := loaders[provider]; ok {
				return loader(ctx)
			}
			return fallbackLoader(ctx)
		},
		memory.NewKeyedExpirableCache[string, meta.Loaded[ps2.WorldsPopulation]](
			len(loaders)+1,
			time.Minute,
		),
	)
	return &populationLoader{
		fallbacks: fallbacks,
		load:      loader.Keyed[string, meta.Loaded[ps2.WorldsPopulation]](cached),
	}
}

func (p *populationLoader) Start(ctx context.Context) {
	p.fallbacks.Start(ctx)
}
