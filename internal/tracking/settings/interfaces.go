package settings

import (
	"context"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type SettingsRepo interface {
	Get(context.Context, discord.ChannelId, ps2_platforms.Platform) (tracking.Settings, error)
	Update(context.Context, discord.ChannelId, ps2_platforms.Platform, tracking.Settings) (tracking.SettingsDiff, error)
}

type ChannelPlatformsRepo interface {
	PlatformsByChannelId(context.Context, discord.ChannelId) ([]ps2_platforms.Platform, error)
}

type OutfitsRepo interface {
	OutfitTagsByIds(context.Context, ps2_platforms.Platform, []ps2.OutfitId) (map[ps2.OutfitId]string, error)
	OutfitIdsByTags(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.OutfitId, error)
}

type CharactersRepo interface {
	CharacterNamesByIds(context.Context, ps2_platforms.Platform, []ps2.CharacterId) (map[ps2.CharacterId]string, error)
	CharacterIdsByNames(context.Context, ps2_platforms.Platform, []string) (map[string]ps2.CharacterId, error)
}

type TrackingRepo interface {
	OnlineOutfitMembers(context.Context, ps2_platforms.Platform, []ps2.OutfitId) (map[ps2.OutfitId]map[ps2.CharacterId]ps2.Character, error)
	OnlineCharacters(context.Context, ps2_platforms.Platform, []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error)
}
