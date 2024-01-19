package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

func startOutfitMembersSynchronizer(
	ctx context.Context,
	// TODO: oss map[string]*outfit_members_synchronizer.OutfitMembersSynchronizer
	//       synchronizers for each platform
	os *outfit_members_synchronizer.OutfitMembersSynchronizer,
	publisher *storage.Publisher,
) error {
	const op = "startOutfitMembersSynchronizer"
	wg := infra.Wg(ctx)
	os.Start(ctx, wg)
	channelOutfitSaved := make(chan storage.ChannelOutfitSaved)
	outfitSavedUnSub, err := publisher.AddHandler(channelOutfitSaved)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer outfitSavedUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case saved := <-channelOutfitSaved:
				if saved.Platform != platforms.PC {
					continue
				}
				os.SyncOutfit(ctx, wg, saved.OutfitId)
			}
		}
	}()
	return nil
}
