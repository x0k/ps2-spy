package sql_storage

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/language"
)

func (storage *Storage) TrackingSettings(ctx context.Context, query discord.SettingsQuery) (discord.TrackingSettings, error) {
	outfits, err := storage.TrackingOutfitIdsForPlatform(ctx, query.ChannelId, query.Platform)
	if err != nil {
		return discord.TrackingSettings{}, err
	}
	characters, err := storage.TrackingCharacterIdsForPlatform(ctx, query.ChannelId, query.Platform)
	if err != nil {
		return discord.TrackingSettings{}, err
	}
	return discord.TrackingSettings{
		Outfits:    outfits,
		Characters: characters,
	}, nil
}

func (storage *Storage) SaveTrackingSettings(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	settings discord.TrackingSettings,
	lang language.Tag,
) error {
	old, err := storage.TrackingSettings(ctx, discord.SettingsQuery{ChannelId: channelId, Platform: platform})
	if err != nil {
		return err
	}
	diff := discord.CalculateTrackingSettingsDiff(old, settings)
	return storage.Begin(
		ctx,
		len(diff.Outfits.ToAdd)+len(diff.Outfits.ToDel)+len(diff.Characters.ToAdd)+len(diff.Characters.ToDel),
		func(tx *Storage) error {
			for _, outfitId := range diff.Outfits.ToDel {
				if err := tx.DeleteChannelOutfit(ctx, channelId, platform, outfitId); err != nil {
					return err
				}
			}
			for _, outfitId := range diff.Outfits.ToAdd {
				if err := tx.SaveChannelOutfit(ctx, channelId, platform, outfitId); err != nil {
					return err
				}
			}
			for _, characterId := range diff.Characters.ToDel {
				if err := tx.DeleteChannelCharacter(ctx, channelId, platform, characterId); err != nil {
					return err
				}
			}
			for _, characterId := range diff.Characters.ToAdd {
				if err := tx.SaveChannelCharacter(ctx, channelId, platform, characterId); err != nil {
					return err
				}
			}
			// Probably this is first settings update
			if len(old.Characters) == 0 && len(old.Outfits) == 0 {
				if err := tx.SaveChannelLanguage(ctx, channelId, lang); err != nil {
					return err
				}
			}
			return nil
		},
	)
}
