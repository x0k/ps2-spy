package ps2_module

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func newOutfitMembersSynchronizerService(
	platform ps2_platforms.Platform,
	outfitMembersSynchronizer *outfit_members_synchronizer.OutfitMembersSynchronizer,
) module.Service {
	return module.NewService(
		fmt.Sprintf("ps2.%s.outfit_members_synchronizer", platform),
		func(ctx context.Context) error {
			outfitMembersSynchronizer.Start(ctx)
			return nil
		},
	)
}
