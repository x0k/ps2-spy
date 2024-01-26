package trackable_online_entities_loader

import (
	"context"
	"errors"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/population_tracker"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrPopulationTrackerNotFound = errors.New("population tracker not found")

type PopulationTrackerLoader struct {
	settingsLoader     loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings]
	populationTrackers map[platforms.Platform]*population_tracker.PopulationTracker
}

func New(
	settingsLoader loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings],
	populationTrackers map[platforms.Platform]*population_tracker.PopulationTracker,
) *PopulationTrackerLoader {
	return &PopulationTrackerLoader{
		settingsLoader:     settingsLoader,
		populationTrackers: populationTrackers,
	}
}

func (l *PopulationTrackerLoader) Load(
	ctx context.Context,
	query meta.SettingsQuery,
) (meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character], error) {
	settings, err := l.settingsLoader.Load(ctx, query)
	if err != nil {
		return meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{}, err
	}
	populationTracker, ok := l.populationTrackers[query.Platform]
	if !ok {
		return meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{}, ErrPopulationTrackerNotFound
	}
	return populationTracker.TrackableOnlineEntities(settings), nil
}
