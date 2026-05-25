package tracking_storage_tracking_repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
)

type Repository struct {
	storage storage.Storage
	log     *slog.Logger
}

func New(storage storage.Storage, log *slog.Logger) *Repository {
	return &Repository{
		storage: storage,
		log:     log,
	}
}

func (r *Repository) TrackingChannelsForCharacter(
	ctx context.Context,
	platform ps2_platforms.Platform,
	characterId ps2.CharacterId,
	outfitId ps2.OutfitId,
) ([]discord.Channel, error) {
	rows, err := r.storage.Queries().ListPlatformTrackingChannelsForCharacter(ctx, db.ListPlatformTrackingChannelsForCharacterParams{
		Platform:    string(platform),
		CharacterID: string(characterId),
		OutfitID:    string(outfitId),
	})
	if err != nil {
		return nil, fmt.Errorf("tracking channels for character %s: %w", characterId, err)
	}
	channels := make([]discord.Channel, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, storage.ChannelFromDTO(ctx, r.log, row))
	}
	return channels, nil
}

func (r *Repository) TrackingChannelsForOutfit(
	ctx context.Context,
	platform ps2_platforms.Platform,
	outfitId ps2.OutfitId,
) ([]discord.Channel, error) {
	rows, err := r.storage.Queries().ListPlatformTrackingChannelsForOutfit(ctx, db.ListPlatformTrackingChannelsForOutfitParams{
		Platform: string(platform),
		OutfitID: string(outfitId),
	})
	if err != nil {
		return nil, fmt.Errorf("tracking channels for outfit %s: %w", outfitId, err)
	}
	channels := make([]discord.Channel, 0, len(rows))
	for _, row := range rows {
		channels = append(channels, storage.ChannelFromDTO(ctx, r.log, row))
	}
	return channels, nil
}

func (r *Repository) AllTrackableCharacterIdsWithDuplicationsForPlatform(
	ctx context.Context, platform ps2_platforms.Platform,
) ([]ps2.CharacterId, error) {
	list, err := r.storage.Queries().ListTrackableCharacterIdsWithDuplicationForPlatform(ctx, string(platform))
	if err != nil {
		return nil, fmt.Errorf("trackable character ids for platform %s: %w", platform, err)
	}
	ids := make([]ps2.CharacterId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.CharacterId(id))
	}
	return ids, nil
}

func (r *Repository) AllTrackableOutfitIdsWithDuplicationsForPlatform(
	ctx context.Context, platform ps2_platforms.Platform,
) ([]ps2.OutfitId, error) {
	list, err := r.storage.Queries().ListTrackableOutfitIdsWithDuplicationForPlatform(ctx, string(platform))
	if err != nil {
		return nil, fmt.Errorf("trackable outfit ids for platform %s: %w", platform, err)
	}
	ids := make([]ps2.OutfitId, 0, len(list))
	for _, id := range list {
		ids = append(ids, ps2.OutfitId(id))
	}
	return ids, nil
}
