package settings

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type Repository struct {
	storage storage.Storage
	log     *logger.Logger
}

func NewRepository(storage storage.Storage, log *logger.Logger) *Repository {
	return &Repository{
		storage: storage,
		log:     log,
	}
}

func (r *Repository) Get(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform) (tracking.Settings, error) {
	channelIdStr := string(channelId)
	platformStr := string(platform)
	characterIds, err := r.storage.Queries().ListChannelCharacterIdsForPlatform(ctx, db.ListChannelCharacterIdsForPlatformParams{
		ChannelID: channelIdStr,
		Platform:  platformStr,
	})
	if err != nil {
		return tracking.Settings{}, fmt.Errorf("failed to list characters: %w", err)
	}
	outfitIds, err := r.storage.Queries().ListChannelOutfitIdsForPlatform(ctx, db.ListChannelOutfitIdsForPlatformParams{
		ChannelID: channelIdStr,
		Platform:  platformStr,
	})
	if err != nil {
		return tracking.Settings{}, fmt.Errorf("failed to list outfits: %w", err)
	}
	return tracking.Settings{
		Characters: slicesx.Map(characterIds, ps2.CharacterIdFromString),
		Outfits:    slicesx.Map(outfitIds, ps2.OutfitIdFromString),
	}, nil
}

func (r *Repository) PlatformsByChannelId(
	ctx context.Context, channelId discord.ChannelId,
) ([]ps2_platforms.Platform, error) {
	data, err := r.storage.Queries().ListChannelTrackablePlatforms(
		ctx, string(channelId),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list channel %q trackable platforms: %w", channelId, err)
	}
	platforms := make([]ps2_platforms.Platform, 0, len(data))
	for _, p := range data {
		platforms = append(platforms, ps2_platforms.Platform(p))
	}
	return platforms, nil
}

func (r *Repository) Update(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	settings tracking.Settings,
) (tracking.SettingsDiff, error) {
	channelIdStr := string(channelId)
	platformStr := string(platform)
	newCharacters := slicesx.Map(
		settings.Characters,
		ps2.CharacterIdToString,
	)
	newOutfits := slicesx.Map(settings.Outfits, ps2.OutfitIdToString)

	var charactersDiff diff.Diff[string]
	var outfitsDiff diff.Diff[string]

	err := r.storage.Transaction(ctx, func(s storage.Storage) error {
		oldCharacters, err := s.Queries().ListChannelCharacterIdsForPlatform(ctx, db.ListChannelCharacterIdsForPlatformParams{
			ChannelID: channelIdStr,
			Platform:  platformStr,
		})
		if err != nil {
			return fmt.Errorf("failed to list characters: %w", err)
		}
		oldOutfits, err := s.Queries().ListChannelOutfitIdsForPlatform(ctx, db.ListChannelOutfitIdsForPlatformParams{
			ChannelID: channelIdStr,
			Platform:  platformStr,
		})
		if err != nil {
			return fmt.Errorf("failed to list outfits: %w", err)
		}

		charactersDiff = diff.SlicesDiff(oldCharacters, newCharacters)
		if err := s.Queries().DeleteChannelCharacters(ctx, db.DeleteChannelCharactersParams{
			ChannelID:    channelIdStr,
			Platform:     platformStr,
			CharacterIds: charactersDiff.ToDel,
		}); err != nil {
			return fmt.Errorf("failed to delete characters: %w", err)
		}

		outfitsDiff = diff.SlicesDiff(oldOutfits, newOutfits)
		if err := s.Queries().DeleteChannelOutfits(ctx, db.DeleteChannelOutfitsParams{
			ChannelID: channelIdStr,
			Platform:  platformStr,
			OutfitIds: outfitsDiff.ToDel,
		}); err != nil {
			return fmt.Errorf("failed to delete outfits: %w", err)
		}

		for _, characterId := range charactersDiff.ToAdd {
			if err := s.Queries().InsertChannelCharacter(ctx, db.InsertChannelCharacterParams{
				ChannelID:   channelIdStr,
				CharacterID: characterId,
				Platform:    platformStr,
			}); err != nil {
				return fmt.Errorf("failed to insert character: %w", err)
			}
		}
		for _, outfitId := range outfitsDiff.ToAdd {
			if err := s.Queries().InsertChannelOutfit(ctx, db.InsertChannelOutfitParams{
				ChannelID: channelIdStr,
				Platform:  platformStr,
				OutfitID:  outfitId,
			}); err != nil {
				return fmt.Errorf("failed to insert outfit: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return tracking.SettingsDiff{}, fmt.Errorf("failed to run transaction: %w", err)
	}
	return tracking.SettingsDiff{
		Characters: diff.Map(charactersDiff, ps2.CharacterIdFromString),
		Outfits:    diff.Map(outfitsDiff, ps2.OutfitIdFromString),
	}, nil
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
		channels = append(channels, storage.ChannelFromDTO(ctx, r.log.Logger, row))
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
		channels = append(channels, storage.ChannelFromDTO(ctx, r.log.Logger, row))
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
