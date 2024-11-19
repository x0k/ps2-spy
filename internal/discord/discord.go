package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type ChannelId string

type TrackableEntities[O any, C any] struct {
	Outfits    O
	Characters C
}

type SubscriptionSettings = TrackableEntities[[]ps2.OutfitId, []ps2.CharacterId]

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
	Platform  ps2_platforms.Platform
}

type PlatformQuery[T any] struct {
	Platform ps2_platforms.Platform
	Value    T
}

type Locale string

const (
	EN Locale = "en"
	RU Locale = "ru"
)

const DEFAULT_LOCALE = EN

func LocaleFromInteraction(i *discordgo.InteractionCreate) Locale {
	switch i.Locale {
	case discordgo.Russian:
		return RU
	case discordgo.EnglishGB:
		return EN
	case discordgo.EnglishUS:
		return EN
	default:
		return DEFAULT_LOCALE
	}
}

type Channel struct {
	ChannelId ChannelId
	Locale    Locale
}

func NewChannel(channelId ChannelId, locale Locale) Channel {
	return Channel{
		ChannelId: channelId,
		Locale:    locale,
	}
}
