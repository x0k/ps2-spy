package subscriptionsettings

import (
	"context"

	"github.com/x0k/ps2-spy/internal/loaders"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/storage/sqlite"
)

type SubscriptionSettingsSaver struct {
	platform string
	storage  *sqlite.Storage
	loader   loaders.KeyedLoader[[2]string, meta.SubscriptionSettings]
}

func New(storage *sqlite.Storage) *SubscriptionSettingsSaver {
	return &SubscriptionSettingsSaver{
		storage: storage,
	}
}

func (s *SubscriptionSettingsSaver) Save(ctx context.Context, channelId string, settings meta.SubscriptionSettings) error {
	old, err := s.loader.Load(ctx, [2]string{channelId, s.platform})
	if err != nil {
		return err
	}
	diff := meta.CalculateSubscriptionSettingsDiff(old, settings)
	storage, err := s.storage.Begin(ctx)
	if err != nil {
		return err
	}
	defer storage.Rollback()
	for _, outfit := range diff.Outfits.ToDel {
		if err := storage.DeleteChannelOutfit(ctx, channelId, s.platform, outfit); err != nil {
			return err
		}
	}
	for _, outfit := range diff.Outfits.ToAdd {
		if err := storage.SaveChannelOutfit(ctx, channelId, s.platform, outfit); err != nil {
			return err
		}
	}
	for _, character := range diff.Characters.ToDel {
		if err := storage.DeleteChannelCharacter(ctx, channelId, s.platform, character); err != nil {
			return err
		}
	}
	for _, character := range diff.Characters.ToAdd {
		if err := storage.SaveChannelCharacter(ctx, channelId, s.platform, character); err != nil {
			return err
		}
	}
	return storage.Commit()
}
