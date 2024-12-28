package characters_tracker_ps2_outfits_repo

import (
	"context"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Repository struct {
	trackers map[ps2_platforms.Platform]*characters_tracker.CharactersTracker
}

func New(
	trackers map[ps2_platforms.Platform]*characters_tracker.CharactersTracker,
) *Repository {
	return &Repository{
		trackers: trackers,
	}
}

func (r *Repository) MembersOnline(
	_ context.Context,
	platform ps2_platforms.Platform,
	outfitIds []ps2.OutfitId,
) (map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character, error) {
	return r.trackers[platform].OutfitMembersOnline(outfitIds), nil
}
