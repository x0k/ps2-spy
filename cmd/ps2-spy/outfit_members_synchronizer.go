package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_sync_at_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_loader"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

var ErrPublisherNotFound = fmt.Errorf("publisher not found")

func newOutfitMembersSynchronizer(
	log *logger.Logger,
	storage *sqlite.Storage,
	censusClient *census2.Client,
	membersSaverPublisher *outfit_members_saver.Publisher,
	platform platforms.Platform,
) *outfit_members_synchronizer.OutfitMembersSynchronizer {
	pLog := log.With(slog.String("platform", string(platform)))
	trackableOutfitsLoader := trackable_outfits_loader.NewStorage(storage, platform)
	outfitSyncAtLoader := outfit_sync_at_loader.NewStorage(storage, platform)
	ns := platforms.PlatformNamespace(platform)
	outfitMembersLoader := outfit_member_ids_loader.NewCensus(censusClient, ns)
	outfitMembersSaver := outfit_members_saver.New(pLog, storage, membersSaverPublisher, platform)
	return outfit_members_synchronizer.New(
		pLog,
		trackableOutfitsLoader,
		outfitSyncAtLoader,
		outfitMembersLoader,
		outfitMembersSaver,
		time.Hour*24,
	)
}

func startNewOutfitMembersSynchronizers(
	ctx context.Context,
	wg *sync.WaitGroup,
	log *logger.Logger,
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	publisher *storage.Publisher,
	outfitMembersSaverPublishers map[platforms.Platform]*outfit_members_saver.Publisher,
) error {
	const op = "startOutfitMembersSynchronizer"
	pcOutfitMembersSaverPublisher, ok := outfitMembersSaverPublishers[platforms.PC]
	if !ok {
		return fmt.Errorf("%s %s: %w", op, platforms.PC, ErrPublisherNotFound)
	}
	pcOutfitMembersSynchronizer := newOutfitMembersSynchronizer(
		log,
		sqlStorage,
		censusClient,
		pcOutfitMembersSaverPublisher,
		platforms.PC,
	)

	ps4euOutfitMembersSaverPublisher, ok := outfitMembersSaverPublishers[platforms.PS4_EU]
	if !ok {
		return fmt.Errorf("%s %s: %w", op, platforms.PS4_EU, ErrPublisherNotFound)
	}
	ps4euOutfitMembersSynchronizer := newOutfitMembersSynchronizer(
		log,
		sqlStorage,
		censusClient,
		ps4euOutfitMembersSaverPublisher,
		platforms.PS4_EU,
	)

	ps4usOutfitMembersSaverPublisher, ok := outfitMembersSaverPublishers[platforms.PS4_US]
	if !ok {
		return fmt.Errorf("%s %s: %w", op, platforms.PS4_US, ErrPublisherNotFound)
	}
	ps4usOutfitMembersSynchronizer := newOutfitMembersSynchronizer(
		log,
		sqlStorage,
		censusClient,
		ps4usOutfitMembersSaverPublisher,
		platforms.PS4_US,
	)

	oss := map[platforms.Platform]*outfit_members_synchronizer.OutfitMembersSynchronizer{
		platforms.PC:     pcOutfitMembersSynchronizer,
		platforms.PS4_EU: ps4euOutfitMembersSynchronizer,
		platforms.PS4_US: ps4usOutfitMembersSynchronizer,
	}
	for _, os := range oss {
		os.Start(ctx, wg)
	}
	channelOutfitSaved := make(chan storage.ChannelOutfitSaved)
	outfitSavedUnSub := publisher.AddChannelOutfitSavedHandler(channelOutfitSaved)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer outfitSavedUnSub()
		for {
			select {
			case <-ctx.Done():
				return
			case saved := <-channelOutfitSaved:
				if os, ok := oss[saved.Platform]; ok {
					os.SyncOutfit(ctx, wg, saved.OutfitId)
				} else {
					log.Warn(ctx, "platform not found", slog.String("platform", string(saved.Platform)))
				}
			}
		}
	}()
	return nil
}
