package tracking_settings_loader

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type SettingsRepo interface {
	Get(context.Context, discord.ChannelId, ps2_platforms.Platform) (tracking.Settings, error)
}

type OutfitsRepo interface {
	OutfitNamesByIds(context.Context, ps2_platforms.Platform, []ps2.OutfitId) ([]string, error)
}

type CharactersRepo interface {
	CharacterNamesByIds(context.Context, ps2_platforms.Platform, []ps2.CharacterId) ([]string, error)
}

type SettingsLoader struct {
	repo           SettingsRepo
	outfitsRepo    OutfitsRepo
	charactersRepo CharactersRepo
}

func New(repo SettingsRepo, outfitsRepo OutfitsRepo, charactersRepo CharactersRepo) *SettingsLoader {
	return &SettingsLoader{
		repo:           repo,
		outfitsRepo:    outfitsRepo,
		charactersRepo: charactersRepo,
	}
}

func (l *SettingsLoader) Load(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform) (tracking.SettingsView, error) {
	settings, err := l.repo.Get(ctx, channelId, platform)
	if err != nil {
		return tracking.SettingsView{}, fmt.Errorf("failed to load settings: %w", err)
	}
	outfits, err := l.outfitsRepo.OutfitNamesByIds(ctx, platform, settings.Outfits)
	if err != nil {
		return tracking.SettingsView{}, fmt.Errorf("failed to load outfits: %w", err)
	}
	characters, err := l.charactersRepo.CharacterNamesByIds(ctx, platform, settings.Characters)
	if err != nil {
		return tracking.SettingsView{}, fmt.Errorf("failed to load characters: %w", err)
	}
	return tracking.SettingsView{
		Outfits:    outfits,
		Characters: characters,
	}, nil
}
