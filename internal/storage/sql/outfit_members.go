package sql_storage

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

func (s *Storage) SaveOutfitMembers(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
	old, err := s.OutfitMembers(ctx, platform, outfitId)
	if err != nil {
		return err
	}
	membersDiff := diff.SlicesDiff(old, members)
	diffSize := len(membersDiff.ToAdd) + len(membersDiff.ToDel)
	if diffSize == 0 {
		return s.SaveOutfitSynchronizedAt(ctx, platform, outfitId, time.Now())
	}
	err = s.Begin(ctx, diffSize, func(tx *Storage) error {
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
	if len(old) == 0 {
		s.publisher.Publish(storage.OutfitMembersInit{
			Platform: platform,
			OutfitId: outfitId,
			Members:  members,
		})
	} else {
		s.publisher.Publish(storage.OutfitMembersUpdate{
			Platform: platform,
			OutfitId: outfitId,
			Members:  membersDiff,
		})
	}
	return s.SaveOutfitSynchronizedAt(ctx, platform, outfitId, time.Now())
}
