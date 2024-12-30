package tracking_settings_data_loader

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/mapx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type SettingsRepo interface {
	Get(context.Context, discord.ChannelId, ps2_platforms.Platform) (tracking.Settings, error)
}

type TrackingRepo interface {
	OnlineOutfitMembers(context.Context, ps2_platforms.Platform, []ps2.OutfitId) (map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character, error)
	OnlineCharacters(context.Context, ps2_platforms.Platform, []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error)
}

type DataLoader struct {
	repo         SettingsRepo
	trackingRepo TrackingRepo
}

func New(repo SettingsRepo, onlineRepo TrackingRepo) *DataLoader {
	return &DataLoader{
		repo:         repo,
		trackingRepo: onlineRepo,
	}
}

func (l *DataLoader) Load(
	ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform,
) (tracking.SettingsData, error) {
	settings, err := l.repo.Get(ctx, channelId, platform)
	if err != nil {
		return tracking.SettingsData{}, fmt.Errorf("failed to load settings: %w", err)
	}
	members, err := l.trackingRepo.OnlineOutfitMembers(ctx, platform, settings.Outfits)
	if err != nil {
		return tracking.SettingsData{}, fmt.Errorf("failed to load outfit members: %w", err)
	}
	characters, err := l.trackingRepo.OnlineCharacters(ctx, platform, settings.Characters)
	if err != nil {
		return tracking.SettingsData{}, fmt.Errorf("failed to load characters: %w", err)
	}
	outfits := make(map[ps2.OutfitId][]ps2.Character, len(members))
	for outfitId, members := range members {
		outfits[outfitId] = mapx.Values(members)
	}
	return tracking.SettingsData{
		Characters: mapx.Values(characters),
		Outfits:    outfits,
	}, nil
}
