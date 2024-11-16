package sql_outfit_members_saver

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sqlite "github.com/x0k/ps2-spy/internal/storage/sql"
)

type OutfitMembersSaver struct {
	log      *logger.Logger
	storage  *sqlite.Storage
	pub      pubsub.Publisher[Event]
	platform platforms.Platform
}

func New(log *logger.Logger, storage *sqlite.Storage, pub pubsub.Publisher[Event], platform platforms.Platform) *OutfitMembersSaver {
	return &OutfitMembersSaver{
		log:      log,
		storage:  storage,
		pub:      pub,
		platform: platform,
	}
}

func (s *OutfitMembersSaver) Save(ctx context.Context, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
	old, err := s.storage.OutfitMembers(ctx, s.platform, outfitId)
	if err != nil {
		return err
	}
	membersDiff := diff.SlicesDiff(old, members)
	diffSize := len(membersDiff.ToAdd) + len(membersDiff.ToDel)
	if diffSize == 0 {
		return s.storage.SaveOutfitSynchronizedAt(ctx, s.platform, outfitId, time.Now())
	}
	err = s.storage.Begin(ctx, diffSize, func(tx *sqlite.Storage) error {
		for _, member := range membersDiff.ToAdd {
			if err := tx.SaveOutfitMember(ctx, s.platform, outfitId, member); err != nil {
				return err
			}
		}
		for _, member := range membersDiff.ToDel {
			if err := tx.DeleteOutfitMember(ctx, s.platform, outfitId, member); err != nil {
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
		pErr = s.pub.Publish(OutfitMembersInit{
			OutfitId: outfitId,
			Members:  members,
		})
	} else {
		pErr = s.pub.Publish(OutfitMembersUpdate{
			OutfitId: outfitId,
			Members:  membersDiff,
		})
	}
	if pErr != nil {
		s.log.Error(ctx, "failed to publish event", sl.Err(pErr))
	}
	return s.storage.SaveOutfitSynchronizedAt(ctx, s.platform, outfitId, time.Now())
}
