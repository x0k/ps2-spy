package sql_subscription_settings_saver

import (
	"context"

	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/meta"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	sql_storage "github.com/x0k/ps2-spy/internal/storage/sql"
)

type SubscriptionSettingsSaver struct {
	platform       ps2_platforms.Platform
	storage        *sql_storage.Storage
	settingsLoader loader.Keyed[meta.SettingsQuery, meta.SubscriptionSettings]
}

func New(
	storage *sql_storage.Storage,
	settingsLoader loader.Keyed[meta.SettingsQuery, meta.SubscriptionSettings],
	platform ps2_platforms.Platform,
) *SubscriptionSettingsSaver {
	return &SubscriptionSettingsSaver{
		storage:        storage,
		platform:       platform,
		settingsLoader: settingsLoader,
	}
}

func (s *SubscriptionSettingsSaver) Save(ctx context.Context, channelId meta.ChannelId, settings meta.SubscriptionSettings) error {
	old, err := s.settingsLoader(ctx, meta.SettingsQuery{ChannelId: channelId, Platform: s.platform})
	if err != nil {
		return err
	}
	diff := meta.CalculateSubscriptionSettingsDiff(old, settings)
	return s.storage.Begin(
		ctx,
		len(diff.Outfits.ToAdd)+len(diff.Outfits.ToDel)+len(diff.Characters.ToAdd)+len(diff.Characters.ToDel),
		func(tx *sql_storage.Storage) error {
			for _, outfitId := range diff.Outfits.ToDel {
				if err := tx.DeleteChannelOutfit(ctx, channelId, s.platform, outfitId); err != nil {
					return err
				}
			}
			for _, outfitId := range diff.Outfits.ToAdd {
				if err := tx.SaveChannelOutfit(ctx, channelId, s.platform, outfitId); err != nil {
					return err
				}
			}
			for _, characterId := range diff.Characters.ToDel {
				if err := tx.DeleteChannelCharacter(ctx, channelId, s.platform, characterId); err != nil {
					return err
				}
			}
			for _, characterId := range diff.Characters.ToAdd {
				if err := tx.SaveChannelCharacter(ctx, channelId, s.platform, characterId); err != nil {
					return err
				}
			}
			return nil
		},
	)
}
