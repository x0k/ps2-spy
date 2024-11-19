package characters_tracker_trackable_online_entities_loader

import (
	"context"
	"errors"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var ErrPopulationTrackerNotFound = errors.New("population tracker not found")

func New(
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.SubscriptionSettings],
	populationTrackers map[ps2_platforms.Platform]*characters_tracker.CharactersTracker,
) loader.Keyed[discord.SettingsQuery, discord.TrackableEntities[
	map[ps2.OutfitId][]ps2.Character,
	[]ps2.Character,
]] {
	return func(
		ctx context.Context,
		query discord.SettingsQuery,
	) (discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character], error) {
		settings, err := settingsLoader(ctx, query)
		if err != nil {
			return discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{}, err
		}
		populationTracker, ok := populationTrackers[query.Platform]
		if !ok {
			return discord.TrackableEntities[map[ps2.OutfitId][]ps2.Character, []ps2.Character]{}, ErrPopulationTrackerNotFound
		}
		return populationTracker.TrackableOnlineEntities(settings), nil
	}
}
