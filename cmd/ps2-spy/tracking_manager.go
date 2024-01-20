package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_character_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_with_duplication_loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/publisher"
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
	// TODO: remove loader
	// outfitTrackersCountLoader := outfit_trackers_count_loader.NewStorage(storage, platform)
	outfitTrackingChannelsLoader := outfit_tracking_channels_loader.NewStorage(storage, platform)
	trackableOutfitsLoader := trackable_outfits_with_duplication_loader.NewStorage(storage, platform)
	return tracking_manager.New(
		characterLoader,
		characterTrackingChannelsLoader,
		trackableCharactersLoader,
		outfitTrackingChannelsLoader,
		trackableOutfitsLoader,
	)
}

func startTrackingManager(
	ctx context.Context,
	tms map[string]*tracking_manager.TrackingManager,
	publisher *publisher.Publisher,
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
	channelOutfitSaved := make(chan storage.ChannelOutfitSaved)
	channelOutfitSavedUnSub, err := publisher.AddHandler(channelOutfitSaved)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	channelOutfitDeleted := make(chan storage.ChannelOutfitDeleted)
	channelOutfitDeletedUnSub, err := publisher.AddHandler(channelOutfitDeleted)
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
		defer channelOutfitSavedUnSub()
		defer channelOutfitDeletedUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-channelCharacterSaved:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn("unknown platform: %s", slog.String("platform", e.Platform))
					continue
				}
				tm.TrackCharacter(e.CharacterId)
			case e := <-channelCharacterDeleted:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", e.Platform))
					continue
				}
				tm.UntrackCharacter(e.CharacterId)
			case e := <-outfitMemberSaved:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", e.Platform))
					continue
				}
				tm.TrackOutfitMember(e.CharacterId, e.OutfitTag)
			case e := <-outfitMemberDeleted:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", e.Platform))
					continue
				}
				tm.UntrackOutfitMember(e.CharacterId, e.OutfitTag)
			case e := <-channelOutfitSaved:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", e.Platform))
					continue
				}
				tm.TrackOutfit(e.OutfitId)
			case e := <-channelOutfitDeleted:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn("unknown platform", slog.String("platform", e.Platform))
					continue
				}
				tm.UntrackOutfit(e.OutfitId)
			}
		}
	}()
	return nil
}
