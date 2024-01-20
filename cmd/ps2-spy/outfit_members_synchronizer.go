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
	"github.com/x0k/ps2-spy/internal/publisher"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

var ErrPublisherNotFound = fmt.Errorf("publisher not found")

func newOutfitMembersSynchronizer(
	storage *sqlite.Storage,
	censusClient *census2.Client,
	membersSaverPublisher publisher.Abstract[publisher.Event],
	platform string,
) (*outfit_members_synchronizer.OutfitMembersSynchronizer, error) {
	const op = "newOutfitMembersSynchronizer"
	trackableOutfitsLoader := trackable_outfits_loader.NewStorage(storage, platform)
	outfitSyncAtLoader := outfit_sync_at_loader.NewStorage(storage, platform)
	ns, err := platforms.PlatformNamespace(platform)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	outfitMembersLoader := outfit_member_ids_loader.NewCensus(censusClient, ns)
	outfitMembersSaver := outfit_members_saver.New(storage, membersSaverPublisher, platform)
	return outfit_members_synchronizer.New(
		trackableOutfitsLoader,
		outfitSyncAtLoader,
		outfitMembersLoader,
		outfitMembersSaver,
		time.Hour*24,
	), nil
}

func startOutfitMembersSynchronizers(
	ctx context.Context,
	sqlStorage *sqlite.Storage,
	censusClient *census2.Client,
	publisher *publisher.Publisher,
	outfitMembersSaverPublishers map[string]publisher.Abstract[publisher.Event],
) error {
	const op = "startOutfitMembersSynchronizer"
	log := infra.OpLogger(ctx, op)
	pcOutfitMembersSaverPublisher, ok := outfitMembersSaverPublishers[platforms.PC]
	if !ok {
		return fmt.Errorf("%s %s: %w", op, platforms.PC, ErrPublisherNotFound)
	}
	pcOutfitMembersSynchronizer, err := newOutfitMembersSynchronizer(
		sqlStorage,
		censusClient,
		pcOutfitMembersSaverPublisher,
		platforms.PC,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4euOutfitMembersSaverPublisher, ok := outfitMembersSaverPublishers[platforms.PS4_EU]
	if !ok {
		return fmt.Errorf("%s %s: %w", op, platforms.PS4_EU, ErrPublisherNotFound)
	}
	ps4euOutfitMembersSynchronizer, err := newOutfitMembersSynchronizer(
		sqlStorage,
		censusClient,
		ps4euOutfitMembersSaverPublisher,
		platforms.PS4_EU,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ps4usOutfitMembersSaverPublisher, ok := outfitMembersSaverPublishers[platforms.PS4_US]
	if !ok {
		return fmt.Errorf("%s %s: %w", op, platforms.PS4_US, ErrPublisherNotFound)
	}
	ps4usOutfitMembersSynchronizer, err := newOutfitMembersSynchronizer(
		sqlStorage,
		censusClient,
		ps4usOutfitMembersSaverPublisher,
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
				os.SyncOutfit(ctx, wg, saved.OutfitTag)
			}
		}
	}()
	return nil
}
