package sql_tracking_settings_saver

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
	"golang.org/x/text/language"
)

func New(
	storage *sql_storage.Storage,
	settingsLoader loader.Keyed[discord.SettingsQuery, discord.TrackingSettings],
) func(context.Context, discord.ChannelId, ps2_platforms.Platform, discord.TrackingSettings, language.Tag) error {
	return func(
		ctx context.Context,
		channelId discord.ChannelId,
		platform ps2_platforms.Platform,
		settings discord.TrackingSettings,
		lang language.Tag,
	) error {
		old, err := settingsLoader(ctx, discord.SettingsQuery{ChannelId: channelId, Platform: platform})
		if err != nil {
			return err
		}
		diff := discord.CalculateTrackingSettingsDiff(old, settings)
		return storage.Begin(
			ctx,
			len(diff.Outfits.ToAdd)+len(diff.Outfits.ToDel)+len(diff.Characters.ToAdd)+len(diff.Characters.ToDel),
			func(tx *sql_storage.Storage) error {
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
}
