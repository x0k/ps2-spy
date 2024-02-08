package meta

import (
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type ChannelId string

type TrackableEntities[O any, C any] struct {
	Outfits    O
	Characters C
}

type SubscriptionSettings TrackableEntities[[]ps2.OutfitId, []ps2.CharacterId]

func CalculateSubscriptionSettingsDiff(
	old SubscriptionSettings,
	new SubscriptionSettings,
) TrackableEntities[diff.Diff[ps2.OutfitId], diff.Diff[ps2.CharacterId]] {
	return TrackableEntities[diff.Diff[ps2.OutfitId], diff.Diff[ps2.CharacterId]]{
		Outfits:    diff.SlicesDiff(old.Outfits, new.Outfits),
		Characters: diff.SlicesDiff(old.Characters, new.Characters),
	}
}

type SettingsQuery struct {
	ChannelId ChannelId
	Platform  platforms.Platform
}

type PlatformQuery[T any] struct {
	Platform platforms.Platform
	Value    T
}
