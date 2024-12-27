package tracking_settings_updater

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/mapx"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type SettingsRepo interface {
	Update(context.Context, discord.ChannelId, ps2_platforms.Platform, tracking.Settings) (tracking.SettingsDiff, error)
}

type OutfitsRepo interface {
	OutfitIdsByTags(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.OutfitId, error)
}

type CharactersRepo interface {
	CharacterIdsByNames(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.CharacterId, error)
}

type SettingsUpdater struct {
	repo                      SettingsRepo
	outfitsRepo               OutfitsRepo
	charactersRepo            CharactersRepo
	maxNumOfTrackedOutfits    int
	maxNumOfTrackedCharacters int
	publisher                 pubsub.Publisher[tracking.Event]
}

func New(
	repo SettingsRepo,
	outfitsRepo OutfitsRepo,
	charactersRepo CharactersRepo,
	maxNumOfTrackedOutfits int,
	maxNumOfTrackedCharacters int,
	publisher pubsub.Publisher[tracking.Event],
) *SettingsUpdater {
	return &SettingsUpdater{
		repo:                      repo,
		outfitsRepo:               outfitsRepo,
		charactersRepo:            charactersRepo,
		maxNumOfTrackedOutfits:    maxNumOfTrackedOutfits,
		maxNumOfTrackedCharacters: maxNumOfTrackedCharacters,
		publisher:                 publisher,
	}
}

func (s *SettingsUpdater) Update(
	ctx context.Context,
	channelId discord.ChannelId,
	platform ps2_platforms.Platform,
	settings tracking.SettingsView,
	updater discord.UserId,
) error {
	if len(settings.Outfits) > s.maxNumOfTrackedOutfits {
		return tracking.ErrTooManyOutfits(settings)
	}
	if len(settings.Characters) > s.maxNumOfTrackedCharacters {
		return tracking.ErrTooManyCharacters(settings)
	}

	outfitIds, _ := s.outfitsRepo.OutfitIdsByTags(ctx, platform, settings.Outfits)
	charIds, _ := s.charactersRepo.CharacterIdsByNames(ctx, platform, settings.Characters)

	if len(settings.Outfits) > len(outfitIds) || len(settings.Characters) > len(charIds) {
		return tracking.ErrFailedToIdentifyEntities{
			OutfitTags:     settings.Outfits,
			FoundOutfitIds: outfitIds,
			CharNames:      settings.Characters,
			FoundCharIds:   charIds,
		}
	}

	settingsDiff, err := s.repo.Update(ctx, channelId, platform, tracking.Settings{
		Characters: mapx.Values(charIds),
		Outfits:    mapx.Values(outfitIds),
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	if !settingsDiff.IsEmpty() {
		s.publisher.Publish(tracking.TrackingSettingsUpdated{
			ChannelId: channelId,
			Platform:  platform,
			Diff:      settingsDiff,
			Updater:   updater,
		})
	}

	return nil
}
