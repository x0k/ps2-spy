package ps2_module

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/characters_tracker"
	"github.com/x0k/ps2-spy/internal/lib/module"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func newCharactersTrackerService(
	ctx context.Context,
	platform ps2_platforms.Platform,
	charactersTracker *characters_tracker.CharactersTracker,
) module.Service {
	return module.NewService(
		fmt.Sprintf("characters_tracker.%s", platform),
		func(ctx context.Context) error {
			charactersTracker.Start(ctx)
			return nil
		},
	)
}
