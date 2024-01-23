package outfit_members_saver

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/publisher"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type OutfitMembersSaver struct {
	storage  *sqlite.Storage
	pub      publisher.Abstract[publisher.Event]
	platform platforms.Platform
}

func New(storage *sqlite.Storage, pub publisher.Abstract[publisher.Event], platform platforms.Platform) *OutfitMembersSaver {
	return &OutfitMembersSaver{
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
	if len(old) == 0 {
		s.pub.Publish(OutfitMembersInit{
			OutfitId: outfitId,
			Members:  members,
		})
	} else {
		s.pub.Publish(OutfitMembersUpdate{
			OutfitId: outfitId,
			Members:  membersDiff,
		})
	}
	return s.storage.SaveOutfitSynchronizedAt(ctx, s.platform, outfitId, time.Now())
}
