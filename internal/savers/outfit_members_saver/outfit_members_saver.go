package outfit_members_saver

import (
	"context"

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

func (s *OutfitMembersSaver) Save(ctx context.Context, outfit string, members []string) error {
	old, err := s.storage.OutfitMembers(ctx, s.platform, outfit)
	if err != nil {
		return err
	}
	membersDiff := diff.SlicesDiff(old, members)
	if len(membersDiff.ToAdd)+len(membersDiff.ToDel) == 0 {
		return nil
	}
	storage, err := s.storage.Begin(ctx, len(membersDiff.ToAdd)+len(membersDiff.ToDel))
	if err != nil {
		return err
	}
	defer storage.Rollback()
	for _, member := range membersDiff.ToAdd {
		if err := storage.SaveOutfitMember(ctx, s.platform, outfit, member); err != nil {
			return err
		}
	}
	for _, member := range membersDiff.ToDel {
		if err := storage.DeleteOutfitMember(ctx, s.platform, outfit, member); err != nil {
			return err
		}
	}
	return storage.Commit()
}
