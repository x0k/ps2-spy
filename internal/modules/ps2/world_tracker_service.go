package ps2_module

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/module"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/worlds_tracker"
)

func newWorldsTrackerService(
	platform ps2_platforms.Platform,
	worldsTracker *worlds_tracker.WorldsTracker,
) module.Service {
	return module.NewService(
		fmt.Sprintf("ps2.%s.worlds_tracker", platform),
		func(ctx context.Context) error {
			worldsTracker.Start(ctx)
			return nil
		},
	)
}
