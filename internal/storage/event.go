package storage

import (
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"golang.org/x/text/language"
)

type EventType string

type Event = pubsub.Event[EventType]

const (
	ChannelLanguageSavedType               EventType = "channel_language_saved"
	ChannelCharacterNotificationsSavedType EventType = "channel_character_notifications_saved"
	ChannelOutfitNotificationsSavedType    EventType = "channel_outfit_notifications_saved"
	ChannelTitleUpdatesSavedType           EventType = "channel_title_updates_saved"
	ChannelDefaultTimezoneSavedType        EventType = "channel_default_timezone_saved"
)

type ChannelLanguageSaved struct {
	ChannelId discord.ChannelId
	Language  language.Tag
}

func (e ChannelLanguageSaved) Type() EventType {
	return ChannelLanguageSavedType
}

type ChannelCharacterNotificationsSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelCharacterNotificationsSaved) Type() EventType {
	return ChannelCharacterNotificationsSavedType
}

type ChannelOutfitNotificationsSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelOutfitNotificationsSaved) Type() EventType {
	return ChannelOutfitNotificationsSavedType
}

type ChannelTitleUpdatesSaved struct {
	ChannelId discord.ChannelId
	Enabled   bool
}

func (e ChannelTitleUpdatesSaved) Type() EventType {
	return ChannelTitleUpdatesSavedType
}

type ChannelDefaultTimezoneSaved struct {
	ChannelId discord.ChannelId
	Location  *time.Location
}

func (e ChannelDefaultTimezoneSaved) Type() EventType {
	return ChannelDefaultTimezoneSavedType
}
