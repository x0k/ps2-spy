package ps2_module

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/module"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/outfit_members_synchronizer"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking_manager"
)

func newStorageEventsSubscriptionService(
	log *logger.Logger,
	ps module.PreStopper,
	tms map[ps2_platforms.Platform]*tracking_manager.TrackingManager,
	oss map[ps2_platforms.Platform]*outfit_members_synchronizer.OutfitMembersSynchronizer,
	subs pubsub.SubscriptionsManager[storage.EventType],
) module.Service {
	outfitMemberSaved := storage.Subscribe[storage.OutfitMemberSaved](ps, subs)
	outfitMemberDeleted := storage.Subscribe[storage.OutfitMemberDeleted](ps, subs)
	channelCharacterSaved := storage.Subscribe[storage.ChannelCharacterSaved](ps, subs)
	channelCharacterDeleted := storage.Subscribe[storage.ChannelCharacterDeleted](ps, subs)
	channelOutfitSaved := storage.Subscribe[storage.ChannelOutfitSaved](ps, subs)
	channelOutfitDeleted := storage.Subscribe[storage.ChannelOutfitDeleted](ps, subs)

	return module.NewService(
		"ps2.storage_events_subscription",
		func(ctx context.Context) error {
			wg := &sync.WaitGroup{}
			for {
				select {
				case <-ctx.Done():
					wg.Wait()
					return nil
				case e := <-outfitMemberSaved:
					tm := tms[e.Platform]
					tm.TrackOutfitMember(e.CharacterId, e.OutfitId)
				case e := <-outfitMemberDeleted:
					tm := tms[e.Platform]
					tm.UntrackOutfitMember(e.CharacterId, e.OutfitId)
				case e := <-channelCharacterSaved:
					tm := tms[e.Platform]
					tm.TrackCharacter(e.CharacterId)
				case e := <-channelCharacterDeleted:
					tm := tms[e.Platform]
					tm.UntrackCharacter(e.CharacterId)
				case e := <-channelOutfitSaved:
					tm := tms[e.Platform]
					if err := tm.TrackOutfit(ctx, e.OutfitId); err != nil {
						log.Error(
							ctx,
							"failed to track outfit",
							slog.String("outfit_id", string(e.OutfitId)),
							slog.String("channel_id", string(e.ChannelId)),
							sl.Err(err),
						)
					}
					os := oss[e.Platform]
					wg.Add(1)
					go func() {
						defer wg.Done()
						os.SyncOutfit(ctx, wg, e.OutfitId)
					}()
				case e := <-channelOutfitDeleted:
					tm := tms[e.Platform]
					if err := tm.UntrackOutfit(ctx, e.OutfitId); err != nil {
						log.Error(
							ctx,
							"failed to untrack outfit",
							slog.String("outfit_id", string(e.OutfitId)),
							slog.String("channel_id", string(e.ChannelId)),
							sl.Err(err),
						)
					}
				}
			}
		},
	)
}
