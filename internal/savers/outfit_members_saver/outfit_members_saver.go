package outfit_members_saver

import (
	"context"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type OutfitMembersSaver struct {
	storage  *sqlite.Storage
	platform string
}

func New(storage *sqlite.Storage, platform string) *OutfitMembersSaver {
	return &OutfitMembersSaver{
		storage:  storage,
		platform: platform,
	}
}

func (s *OutfitMembersSaver) Save(ctx context.Context, outfitTag string, members []string) error {
	old, err := s.storage.OutfitMembers(ctx, s.platform, outfitTag)
	if err != nil {
		return err
	}
	membersDiff := diff.SlicesDiff(old, members)
	diffSize := len(membersDiff.ToAdd) + len(membersDiff.ToDel)
	if diffSize == 0 {
		return s.storage.SaveOutfitSynchronizedAt(ctx, s.platform, outfitTag, time.Now())
	}
	err = s.storage.Begin(ctx, diffSize, func(tx *sqlite.Storage) error {
		for _, member := range membersDiff.ToAdd {
			if err := tx.SaveOutfitMember(ctx, s.platform, outfitTag, member); err != nil {
				return err
			}
		}
		for _, member := range membersDiff.ToDel {
			if err := tx.DeleteOutfitMember(ctx, s.platform, outfitTag, member); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.storage.SaveOutfitSynchronizedAt(ctx, s.platform, outfitTag, time.Now())
}
