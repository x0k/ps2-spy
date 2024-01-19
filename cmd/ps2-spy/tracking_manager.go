package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_trackers_count_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_character_ids_loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func newTrackingManager(
	storage *sqlite.Storage,
	characterLoader loaders.KeyedLoader[string, ps2.Character],
	characterTrackingChannelsLoader loaders.KeyedLoader[ps2.Character, []string],
	platform string,
) *tracking_manager.TrackingManager {
	trackableCharactersLoader := trackable_character_ids_loader.NewStorage(storage, platform)
	outfitTrackersCountLoader := outfit_trackers_count_loader.NewStorage(storage, platform)
	return tracking_manager.New(
		characterLoader,
		characterTrackingChannelsLoader,
		trackableCharactersLoader,
		outfitTrackersCountLoader,
	)
}

func startTrackingManager(
	ctx context.Context,
	tms map[string]*tracking_manager.TrackingManager,
	publisher *storage.Publisher,
) error {
	const op = "startTrackingManager"
	log := infra.OpLogger(ctx, op)
	wg := infra.Wg(ctx)
	for _, tm := range tms {
		tm.Start(ctx, wg)
	}
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
				tm, ok := tms[saved.Platform]
				if !ok {
					log.Warn("unknown platform: %s", slog.String("platform", saved.Platform))
					continue
				}
				tm.TrackCharacter(saved.CharacterId)
			case deleted := <-channelCharacterDeleted:
				tm, ok := tms[deleted.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", deleted.Platform))
					continue
				}
				tm.UntrackCharacter(deleted.CharacterId)
			case saved := <-outfitMemberSaved:
				tm, ok := tms[saved.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", saved.Platform))
					continue
				}
				tm.TrackOutfitMember(ctx, saved.CharacterId, saved.OutfitTag)
			case deleted := <-outfitMemberDeleted:
				tm, ok := tms[deleted.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", deleted.Platform))
					continue
				}
				tm.UntrackOutfitMember(ctx, deleted.CharacterId, deleted.OutfitTag)
			}
		}
	}()
	return nil
}
