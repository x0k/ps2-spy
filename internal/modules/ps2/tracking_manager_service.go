package ps2_module

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/module"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func newTrackingManagerService(
	platform ps2_platforms.Platform,
	trackingManager *tracking_manager.TrackingManager,
) module.Service {
	return module.NewService(
		fmt.Sprintf("ps2.%s.tracking_manager", platform),
		func(ctx context.Context) error {
			trackingManager.Start(ctx)
			return nil
		},
	)
}
