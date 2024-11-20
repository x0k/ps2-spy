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

type worldPopulationLoader struct {
	fallbacks *containers.Fallbacks[loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]]]
	load      loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]]
}

func newWorldPopulationLoader(
	log *logger.Logger,
	loaders map[string]loader.Keyed[ps2.WorldId, meta.Loaded[ps2.DetailedWorldPopulation]],
	loadersPriority []string,
) *worldPopulationLoader {
	fallbacks := containers.NewFallbacks(
		log.Logger.With(sl.Component("world_population_loader_fallbacks")),
		loaders,
		loadersPriority,
		time.Hour,
	)
	fallbackLoader := loader.NewKeyedFallback(fallbacks)
	cached := loader.WithQueriedCache(
		log.Logger.With(sl.Component("world_population_loader_cache")),
		func(ctx context.Context, query query[ps2.WorldId]) (meta.Loaded[ps2.DetailedWorldPopulation], error) {
			if loader, ok := loaders[query.Provider]; ok {
				return loader(ctx, query.Key)
			}
			return fallbackLoader(ctx, query.Key)
		},
		memory.NewKeyedExpirableCache[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]](
			(len(loaders)+1)*len(ps2.ZoneNames),
			time.Minute,
		),
	)
	return &worldPopulationLoader{
		fallbacks: fallbacks,
		load:      loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]](cached),
	}
}

func (l *worldPopulationLoader) Start(ctx context.Context) {
	l.fallbacks.Start(ctx)
}
