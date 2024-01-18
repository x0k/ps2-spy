package main

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

func startOutfitMembersSynchronizer(
	s *setup,
	// TODO: oss map[string]*outfit_members_synchronizer.OutfitMembersSynchronizer
	//       synchronizers for each platform
	os *outfit_members_synchronizer.OutfitMembersSynchronizer,
	publisher *storage.Publisher,
) error {
	const op = "startOutfitMembersSynchronizer"
	os.Start(s.ctx, s.wg)
	channelOutfitSaved := make(chan storage.ChannelOutfitSaved)
	outfitSavedUnSub, err := publisher.AddHandler(channelOutfitSaved)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer outfitSavedUnSub()
		for {
			select {
			case <-s.ctx.Done():
				return
			case saved := <-channelOutfitSaved:
				if saved.Platform != platforms.PC {
					continue
				}
				os.SyncOutfit(s.ctx, s.wg, saved.OutfitId)
			}
		}
	}()
	return nil
}
