package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_member_ids_loader"
	"github.com/x0k/ps2-spy/internal/loaders/outfit_sync_at_loader"
	"github.com/x0k/ps2-spy/internal/loaders/trackable_outfits_loader"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

func newOutfitMembersSynchronizer(
	storage *sqlite.Storage,
	censusClient *census2.Client,
	platform string,
) (*outfit_members_synchronizer.OutfitMembersSynchronizer, error) {
	trackableOutfitsLoader := trackable_outfits_loader.NewStorage(storage, platform)
	outfitSyncAtLoader := outfit_sync_at_loader.NewStorage(storage, platform)
	outfitMembersLoader, err := outfit_member_ids_loader.NewCensus(censusClient, platform)
	if err != nil {
		return nil, err
	}
	outfitMembersSaver := outfit_members_saver.New(storage, platform)
	return outfit_members_synchronizer.New(
		trackableOutfitsLoader,
		outfitSyncAtLoader,
		outfitMembersLoader,
		outfitMembersSaver,
		time.Hour*24,
	), nil
}

func startOutfitMembersSynchronizer(
	ctx context.Context,
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	publisher *storage.Publisher,
) error {
	const op = "startOutfitMembersSynchronizer"
	log := infra.OpLogger(ctx, op)
	pcOutfitMembersSynchronizer, err := newOutfitMembersSynchronizer(
		sqlStorage,
		censusClient,
		platforms.PC,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ps4euOutfitMembersSynchronizer, err := newOutfitMembersSynchronizer(
		sqlStorage,
		censusClient,
		platforms.PS4_EU,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	ps4usOutfitMembersSynchronizer, err := newOutfitMembersSynchronizer(
		sqlStorage,
		censusClient,
		platforms.PS4_US,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	oss := map[string]*outfit_members_synchronizer.OutfitMembersSynchronizer{
		platforms.PC:     pcOutfitMembersSynchronizer,
		platforms.PS4_EU: ps4euOutfitMembersSynchronizer,
		platforms.PS4_US: ps4usOutfitMembersSynchronizer,
	}
	wg := infra.Wg(ctx)
	for _, os := range oss {
		os.Start(ctx, wg)
	}
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
				os, ok := oss[saved.Platform]
				if !ok {
					log.Warn("platform not found", slog.String("platform", saved.Platform))
					continue
				}
				os.SyncOutfit(ctx, wg, saved.OutfitId)
			}
		}
	}()
	return nil
}
