package storage_tracking_settings_repo

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/storage"
	"github.com/x0k/ps2-spy/internal/tracking"
)

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
