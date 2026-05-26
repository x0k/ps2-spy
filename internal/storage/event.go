package storage

import (
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/language"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	ChannelLanguageSavedType         EventType = "channel_language_saved"
	ChannelTitleUpdatesSavedType     EventType = "channel_title_updates_saved"
	ChannelTrackingSettingsSavedType EventType = "channel_tracking_settings_saved"
)

type ChannelLanguageSaved struct {
	ChannelId discord.ChannelId
	Language  language.Tag
}

func (e ChannelLanguageSaved) Type() EventType {
	return ChannelLanguageSavedType
}

type ChannelTitleUpdatesSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelTitleUpdatesSaved) Type() EventType {
	return ChannelTitleUpdatesSavedType
}

type ChannelTrackingSettingsSaved struct {
	ChannelId discord.ChannelId
	Platform  ps2_platforms.Platform
	Updater   discord.UserId
}

func (e ChannelTrackingSettingsSaved) Type() EventType {
	return ChannelTrackingSettingsSavedType
}
