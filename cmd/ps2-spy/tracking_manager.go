package main

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func startTrackingManager(
	ctx context.Context,
	// TODO: tms map[string]*tracking_manager.TrackingManager
	//       managers for each platform
	pcTm *tracking_manager.TrackingManager,
	publisher *storage.Publisher,
) error {
	const op = "startTrackingManager"
	wg := infra.Wg(ctx)
	pcTm.Start(ctx, wg)
	channelCharacterSaved := make(chan storage.ChannelCharacterSaved)
	charSavedUnSub, err := publisher.AddHandler(channelCharacterSaved)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	channelCharacterDeleted := make(chan storage.ChannelCharacterDeleted)
	charDeletedUnSub, err := publisher.AddHandler(channelCharacterDeleted)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	outfitMemberSaved := make(chan storage.OutfitMemberSaved)
	outfitMemberSavedUnSub, err := publisher.AddHandler(outfitMemberSaved)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	outfitMemberDeleted := make(chan storage.OutfitMemberDeleted)
	outfitMemberDeletedUnSub, err := publisher.AddHandler(outfitMemberDeleted)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer charSavedUnSub()
		defer charDeletedUnSub()
		defer outfitMemberSavedUnSub()
		defer outfitMemberDeletedUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case saved := <-channelCharacterSaved:
				if saved.Platform != platforms.PC {
					continue
				}
				pcTm.TrackCharacter(saved.CharacterId)
			case deleted := <-channelCharacterDeleted:
				if deleted.Platform != platforms.PC {
					continue
				}
				pcTm.UntrackCharacter(deleted.CharacterId)
			case saved := <-outfitMemberSaved:
				if saved.Platform != platforms.PC {
					continue
				}
				pcTm.TrackOutfitMember(ctx, saved.CharacterId, saved.OutfitTag)
			case deleted := <-outfitMemberDeleted:
				if deleted.Platform != platforms.PC {
					continue
				}
				pcTm.UntrackOutfitMember(ctx, deleted.CharacterId, deleted.OutfitTag)
			}
		}
	}()
	return nil
}
