package main

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_tracking_channels_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_character_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_with_duplication_loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func newTrackingManager(
	log *logger.Logger,
	storage *sqlite.Storage,
	characterLoader loaders.KeyedLoader[ps2.CharacterId, ps2.Character],
	characterTrackingChannelsLoader loaders.KeyedLoader[ps2.Character, []meta.ChannelId],
	platform platforms.Platform,
) *tracking_manager.TrackingManager {
	trackableCharactersLoader := trackable_character_ids_loader.NewStorage(storage, platform)
	outfitMembersLoader := outfit_member_ids_loader.NewStorage(storage, platform)
	outfitTrackingChannelsLoader := outfit_tracking_channels_loader.NewStorage(storage, platform)
	trackableOutfitsLoader := trackable_outfits_with_duplication_loader.NewStorage(storage, platform)
	return tracking_manager.New(
		log,
		characterLoader,
		characterTrackingChannelsLoader,
		trackableCharactersLoader,
		outfitMembersLoader,
		outfitTrackingChannelsLoader,
		trackableOutfitsLoader,
	)
}

func startTrackingManager(
	ctx context.Context,
	log *logger.Logger,
	tms map[platforms.Platform]*tracking_manager.TrackingManager,
	publisher *storage.Publisher,
) {
	wg := infra.Wg(ctx)
	for _, tm := range tms {
		tm.Start(ctx, wg)
	}
	channelCharacterSaved := make(chan storage.ChannelCharacterSaved)
	charSavedUnSub := publisher.AddChannelCharacterSavedHandler(channelCharacterSaved)
	channelCharacterDeleted := make(chan storage.ChannelCharacterDeleted)
	charDeletedUnSub := publisher.AddChannelCharacterDeletedHandler(channelCharacterDeleted)
	outfitMemberSaved := make(chan storage.OutfitMemberSaved)
	outfitMemberSavedUnSub := publisher.AddOutfitMemberSavedHandler(outfitMemberSaved)
	outfitMemberDeleted := make(chan storage.OutfitMemberDeleted)
	outfitMemberDeletedUnSub := publisher.AddOutfitMemberDeletedHandler(outfitMemberDeleted)
	channelOutfitSaved := make(chan storage.ChannelOutfitSaved)
	channelOutfitSavedUnSub := publisher.AddChannelOutfitSavedHandler(channelOutfitSaved)
	channelOutfitDeleted := make(chan storage.ChannelOutfitDeleted)
	channelOutfitDeletedUnSub := publisher.AddChannelOutfitDeletedHandler(channelOutfitDeleted)
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
					log.Warn(ctx, "unknown platform: %s", slog.String("platform", string(e.Platform)))
					continue
				}
				tm.TrackCharacter(e.CharacterId)
			case e := <-channelCharacterDeleted:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn(ctx, "unknown platform", slog.String("platform", string(e.Platform)))
					continue
				}
				tm.UntrackCharacter(e.CharacterId)
			case e := <-outfitMemberSaved:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn(ctx, "unknown platform", slog.String("platform", string(e.Platform)))
					continue
				}
				tm.TrackOutfitMember(e.CharacterId, e.OutfitId)
			case e := <-outfitMemberDeleted:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn(ctx, "unknown platform", slog.String("platform", string(e.Platform)))
					continue
				}
				tm.UntrackOutfitMember(e.CharacterId, e.OutfitId)
			case e := <-channelOutfitSaved:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn(ctx, "unknown platform", slog.String("platform", string(e.Platform)))
					continue
				}
				err := tm.TrackOutfit(ctx, e.OutfitId)
				if err != nil {
					log.Error(
						ctx,
						"failed to track outfit",
						slog.String("platform", string(e.Platform)),
						slog.String("outfit_id", string(e.OutfitId)),
						sl.Err(err),
					)
				}
			case e := <-channelOutfitDeleted:
				tm, ok := tms[e.Platform]
				if !ok {
					log.Warn(ctx, "unknown platform", slog.String("platform", string(e.Platform)))
					continue
				}
				err := tm.UntrackOutfit(ctx, e.OutfitId)
				if err != nil {
					log.Error(
						ctx,
						"failed to untrack outfit",
						slog.String("platform", string(e.Platform)),
						slog.String("outfit_id", string(e.OutfitId)),
						sl.Err(err),
					)
				}
			}
		}
	}()
}
