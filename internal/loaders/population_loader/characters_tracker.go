package population_loader

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CharactersTrackerLoader struct {
	log                *logger.Logger
	botName            string
	charactersTrackers map[platforms.Platform]*characters_tracker.CharactersTracker
}

func NewCharactersTrackerLoader(
	log *logger.Logger,
	botName string,
	charactersTrackers map[platforms.Platform]*characters_tracker.CharactersTracker,
) *CharactersTrackerLoader {
	return &CharactersTrackerLoader{
		log: log.With(
			slog.String("component", "loaders.population_loader.CharactersTrackerLoader"),
		),
		botName:            botName,
		charactersTrackers: charactersTrackers,
	}
}

func (l *CharactersTrackerLoader) Load(ctx context.Context) (loaders.Loaded[ps2.WorldsPopulation], error) {
	total := 0
	worlds := make([]ps2.WorldPopulation, 0)
	for _, platform := range platforms.Platforms {
		tracker, ok := l.charactersTrackers[platform]
		if !ok {
			l.log.Warn(ctx, "no population tracker for platform", slog.String("platform", string(platform)))
			continue
		}
		population := tracker.WorldsPopulation()
		total += population.Total
		worlds = append(worlds, population.Worlds...)
	}
	return loaders.LoadedNow(l.botName, ps2.WorldsPopulation{
		Total:  total,
		Worlds: worlds,
	}), nil
}
