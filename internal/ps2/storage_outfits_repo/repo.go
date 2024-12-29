package ps2_storage_outfits_repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
	"github.com/x0k/ps2-spy/internal/storage"
)

type Repository struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Repository {
	return &Repository{
		storage: storage,
	}
}

func (r *Repository) Transaction(ctx context.Context, fn func(r *Repository) error) error {
	return r.storage.Transaction(ctx, func(s storage.Storage) error {
		return fn(New(s))
	})
}

func (r *Repository) SynchronizedAt(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId,
) (time.Time, error) {
	time, err := r.storage.Queries().GetPlatformOutfitSynchronizedAt(ctx, db.GetPlatformOutfitSynchronizedAtParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return time, shared.ErrNotFound
	} else if err != nil {
		return time, fmt.Errorf("failed to get outfit sync time: %w", err)
	}
	return time, nil
}

func (r *Repository) SaveSynchronizedAt(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, time time.Time,
) error {
	return r.storage.Queries().UpsertPlatformOutfitSynchronizedAt(ctx, db.UpsertPlatformOutfitSynchronizedAtParams{
		Platform:       string(platform),
		OutfitID:       string(outfitId),
		SynchronizedAt: time,
	})
}

func (r *Repository) TrackableOutfitIds(
	ctx context.Context, platform ps2_platforms.Platform,
) ([]ps2.OutfitId, error) {
	data, err := r.storage.Queries().ListUniqueTrackableOutfitIdsForPlatform(ctx, string(platform))
	if err != nil {
		return nil, fmt.Errorf("failed to list outfits: %w", err)
	}
	outfitIds := make([]ps2.OutfitId, 0, len(data))
	for _, id := range data {
		outfitIds = append(outfitIds, ps2.OutfitId(id))
	}
	return outfitIds, nil
}

func (r *Repository) MemberIds(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId,
) ([]ps2.CharacterId, error) {
	data, err := r.storage.Queries().ListPlatformOutfitMembers(ctx, db.ListPlatformOutfitMembersParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	characterIds := make([]ps2.CharacterId, 0, len(data))
	for _, id := range data {
		characterIds = append(characterIds, ps2.CharacterId(id))
	}
	return characterIds, nil
}

func (r *Repository) AddMember(
	ctx context.Context,
	platform ps2_platforms.Platform,
	outfitId ps2.OutfitId,
	member ps2.CharacterId,
) error {
	return r.storage.Queries().InsertOutfitMember(ctx, db.InsertOutfitMemberParams{
		OutfitID:    string(outfitId),
		CharacterID: string(member),
		Platform:    string(platform),
	})
}

func (r *Repository) RemoveMembers(
	ctx context.Context,
	platform ps2_platforms.Platform,
	outfitId ps2.OutfitId,
	members []ps2.CharacterId,
) error {
	ids := make([]string, 0, len(members))
	for _, id := range members {
		ids = append(ids, string(id))
	}
	return r.storage.Queries().DeleteOutfitMembers(ctx, db.DeleteOutfitMembersParams{
		OutfitID:     string(outfitId),
		CharacterIds: ids,
		Platform:     string(platform),
	})
}
