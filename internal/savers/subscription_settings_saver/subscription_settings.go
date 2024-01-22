package subscription_settings_saver

import (
	"context"

	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type SubscriptionSettingsSaver struct {
	platform       string
	storage        *sqlite.Storage
	settingsLoader loaders.KeyedLoader[[2]string, meta.SubscriptionSettings]
}

func New(
	storage *sqlite.Storage,
	settingsLoader loaders.KeyedLoader[[2]string, meta.SubscriptionSettings],
	platform string,
) *SubscriptionSettingsSaver {
	return &SubscriptionSettingsSaver{
		storage:        storage,
		platform:       platform,
		settingsLoader: settingsLoader,
	}
}

func (s *SubscriptionSettingsSaver) Save(ctx context.Context, channelId string, settings meta.SubscriptionSettings) error {
	old, err := s.settingsLoader.Load(ctx, [2]string{channelId, s.platform})
	if err != nil {
		return err
	}
	diff := meta.CalculateSubscriptionSettingsDiff(old, settings)
	return s.storage.Begin(
		ctx,
		len(diff.Outfits.ToAdd)+len(diff.Outfits.ToDel)+len(diff.Characters.ToAdd)+len(diff.Characters.ToDel),
		func(tx *sqlite.Storage) error {
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
