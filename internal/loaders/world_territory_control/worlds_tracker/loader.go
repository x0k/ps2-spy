package worlds_tracker_world_territory_control_loader

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func New(
	appName string,
	worldsTrackers map[ps2_platforms.Platform]*worlds_tracker.WorldsTracker,
) loader.Keyed[ps2.WorldId, meta.Loaded[ps2.WorldTerritoryControl]] {
	return func(ctx context.Context, worldId ps2.WorldId) (meta.Loaded[ps2.WorldTerritoryControl], error) {
		const op = "worlds_tracker_world_territory_control_loader"
		platform, ok := ps2.WorldPlatforms[worldId]
		if !ok {
			return meta.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("%s: unknown world %q", op, worldId)
		}
		tracker, ok := worldsTrackers[platform]
		if !ok {
			return meta.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("%s: no worlds tracker for platform %s", op, platform)
		}
		territoryControl, err := tracker.WorldTerritoryControl(ctx, worldId)
		if err != nil {
			return meta.Loaded[ps2.WorldTerritoryControl]{}, fmt.Errorf("%s: getting territory control: %w", op, err)
		}
		return meta.LoadedNow(appName, territoryControl), nil
	}
}
