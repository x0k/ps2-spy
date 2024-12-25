package tracking_settings_repo

import (
	"context"
	"fmt"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

func (r *Repository) Get(ctx context.Context, channelId discord.ChannelId, platform ps2_platforms.Platform) (tracking.Settings, error) {
	channelIdStr := string(channelId)
	platformStr := string(platform)
	characterIds, err := r.storage.Queries().ListChannelCharacterIdsForPlatform(ctx, db.ListChannelCharacterIdsForPlatformParams{
		ChannelID: channelIdStr,
		Platform:  platformStr,
	})
	if err != nil {
		return tracking.Settings{}, fmt.Errorf("failed to list characters: %w", err)
	}
	outfitIds, err := r.storage.Queries().ListChannelOutfitIdsForPlatform(ctx, db.ListChannelOutfitIdsForPlatformParams{
		ChannelID: channelIdStr,
		Platform:  platformStr,
	})
	if err != nil {
		return tracking.Settings{}, fmt.Errorf("failed to list outfits: %w", err)
	}
	return tracking.Settings{
		Characters: slicesx.Map(characterIds, ps2.CharacterIdFromString),
		Outfits:    slicesx.Map(outfitIds, ps2.OutfitIdFromString),
	}, nil
}
