package tracking

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Manager struct {
	managers map[ps2_platforms.Platform]*platformManager
	wg       sync.WaitGroup
}

func New(
	log *logger.Logger,
	charLoader CharacterLoader,
	characterTrackingChannelsLoader CharacterTrackingChannelsLoader,
	trackableCharactersLoader TrackableCharactersLoader,
	outfitMembersLoader OutfitMembersLoader,
	outfitTrackingChannelsLoader OutfitTrackingChannelsLoader,
	trackableOutfitsLoader TrackableOutfitsLoader,
) *Manager {
	managers := make(map[ps2_platforms.Platform]*platformManager, len(ps2_platforms.Platforms))
	for _, platform := range ps2_platforms.Platforms {
		managers[platform] = newPlatformManager(
			log.With(slog.String("platform", string(platform))),
			platform,
			charLoader,
			characterTrackingChannelsLoader,
			trackableCharactersLoader,
			outfitMembersLoader,
			outfitTrackingChannelsLoader,
			trackableOutfitsLoader,
		)
	}
	return &Manager{
		managers: managers,
	}
}

func (m *Manager) Start(ctx context.Context) {
	m.wg.Add(len(m.managers))
	for _, manager := range m.managers {
		go func() {
			defer m.wg.Done()
			manager.Start(ctx)
		}()
	}
	<-ctx.Done()
	m.wg.Wait()
}

func (m *Manager) CharacterChannels(
	ctx context.Context, platform ps2_platforms.Platform, characterId ps2.CharacterId,
) ([]discord.Channel, error) {
	manager, err := m.platformManager(platform)
	if err != nil {
		return nil, err
	}
	return manager.ChannelsForCharacter(ctx, characterId)
}

func (m *Manager) OutfitChannels(
	ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId,
) ([]discord.Channel, error) {
	manager, err := m.platformManager(platform)
	if err != nil {
		return nil, err
	}
	return manager.ChannelsForOutfit(ctx, outfitId)
}

func (m *Manager) TrackOutfitMembers(
	outfitId ps2.OutfitId, platform ps2_platforms.Platform, charIds []ps2.CharacterId,
) error {
	manager, err := m.platformManager(platform)
	if err != nil {
		return err
	}
	manager.TrackOutfitMembers(outfitId, charIds)
	return nil
}

func (m *Manager) UntrackOutfitMembers(
	outfitId ps2.OutfitId, platform ps2_platforms.Platform, charIds []ps2.CharacterId,
) error {
	manager, err := m.platformManager(platform)
	if err != nil {
		return err
	}
	manager.UntrackOutfitMembers(outfitId, charIds)
	return nil
}

func (m *Manager) HandleTrackingSettingsUpdate(
	ctx context.Context, platform ps2_platforms.Platform, update TrackingSettingsUpdated,
) error {
	manager, err := m.platformManager(platform)
	if err != nil {
		return err
	}
	manager.HandleTrackingSettingsUpdate(ctx, update)
	return nil
}

func (m *Manager) platformManager(platform ps2_platforms.Platform) (*platformManager, error) {
	manager, ok := m.managers[platform]
	if !ok {
		return nil, ps2_platforms.ErrUnknownPlatform{Platform: platform}
	}
	return manager, nil
}
