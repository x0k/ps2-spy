package sql_outfit_members_saver

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sqlite "github.com/x0k/ps2-spy/internal/storage/sql"
)

func New(
	log *logger.Logger,
	storage *sqlite.Storage,
	pub pubsub.Publisher[Event],
	platform ps2_platforms.Platform,
) func(ctx context.Context, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
	return func(ctx context.Context, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
		old, err := storage.OutfitMembers(ctx, platform, outfitId)
		if err != nil {
			return err
		}
		membersDiff := diff.SlicesDiff(old, members)
		diffSize := len(membersDiff.ToAdd) + len(membersDiff.ToDel)
		if diffSize == 0 {
			return storage.SaveOutfitSynchronizedAt(ctx, platform, outfitId, time.Now())
		}
		err = storage.Begin(ctx, diffSize, func(tx *sqlite.Storage) error {
			for _, member := range membersDiff.ToAdd {
				if err := tx.SaveOutfitMember(ctx, platform, outfitId, member); err != nil {
					return err
				}
			}
			for _, member := range membersDiff.ToDel {
				if err := tx.DeleteOutfitMember(ctx, platform, outfitId, member); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		var pErr error
		if len(old) == 0 {
			pErr = pub.Publish(OutfitMembersInit{
				OutfitId: outfitId,
				Members:  members,
			})
		} else {
			pErr = pub.Publish(OutfitMembersUpdate{
				OutfitId: outfitId,
				Members:  membersDiff,
			})
		}
		if pErr != nil {
			log.Error(ctx, "failed to publish event", sl.Err(pErr))
		}
		return storage.SaveOutfitSynchronizedAt(ctx, platform, outfitId, time.Now())
	}
}
