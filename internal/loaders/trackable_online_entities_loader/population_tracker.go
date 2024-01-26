package trackable_online_entities_loader

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/population_tracker"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type PopulationTrackerLoader struct {
	settingsLoader     loaders.KeyedLoader[meta.SettingsQuery, meta.SubscriptionSettings]
	populationTrackers map[platforms.Platform]*population_tracker.PopulationTracker
}

func (l *PopulationTrackerLoader) Load(
	ctx context.Context,
	channelId meta.ChannelId,
) (meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character], error) {
	outfits := make(map[ps2.OutfitId][]ps2.Character)
	characters := make([]ps2.Character, 0)
	for _, platform := range platforms.Platforms {
		settings, err := l.settingsLoader.Load(ctx, meta.SettingsQuery{
			ChannelId: channelId,
			Platform:  platform,
		})
		if err != nil {
			return meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{}, err
		}
		populationTracker, ok := l.populationTrackers[platform]
		if !ok {
			continue
		}
		entities := populationTracker.TrackableOnlineEntities(settings)
		for id, outfitCharacters := range entities.Outfits {
			outfits[id] = append(outfits[id], outfitCharacters...)
		}
		characters = append(characters, entities.Characters...)
	}
	return meta.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{
		Outfits:    outfits,
		Characters: characters,
	}, nil
}
