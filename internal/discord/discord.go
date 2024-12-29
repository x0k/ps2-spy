package discord

import (
	"time"

	"golang.org/x/text/language"

	"github.com/bwmarrin/discordgo"
)

type ChannelId string

type UserId string

func MemberOrUserId(i *discordgo.InteractionCreate) UserId {
	if i.Member != nil {
		return UserId(i.Member.User.ID)
	}
	return UserId(i.User.ID)
}

type ChannelAndUserIds string

const idsSeparator = "+"

func NewChannelAndUserId(channelId ChannelId, userId UserId) ChannelAndUserIds {
	return ChannelAndUserIds(string(channelId) + idsSeparator + string(userId))
}

var DEFAULT_LANG_TAG = language.English

func UserLocale(i *discordgo.InteractionCreate) language.Tag {
	if t, err := language.Parse(string(i.Locale)); err == nil {
		return t
	}
	return DEFAULT_LANG_TAG
}

func ChannelLocaleOrDefaultToUser(i *discordgo.InteractionCreate) language.Tag {
	if t, err := language.Parse(string(*i.GuildLocale)); err == nil {
		return t
	}
	return UserLocale(i)
}

type Channel struct {
	Id                     ChannelId
	Locale                 language.Tag
	CharacterNotifications bool
	OutfitNotifications    bool
	TitleUpdates           bool
	DefaultTimezone        *time.Location
}

func NewChannel(
	channelId ChannelId,
	locale language.Tag,
	characterNotifications bool,
	outfitNotifications bool,
	titleUpdates bool,
	defaultTimezone *time.Location,
) Channel {
	return Channel{
		Id:                     channelId,
		Locale:                 locale,
		CharacterNotifications: characterNotifications,
		OutfitNotifications:    outfitNotifications,
		TitleUpdates:           titleUpdates,
		DefaultTimezone:        defaultTimezone,
	}
}

func NewDefaultChannel(channelId ChannelId) Channel {
	return NewChannel(channelId, DEFAULT_LANG_TAG, true, true, true, time.UTC)
}

type FormState[T any] struct {
	SubmitButtonId string
	Data           T
}

func IsChannelsManagerOrDM(i *discordgo.InteractionCreate) bool {
	return i.Member == nil || i.Member.Permissions&discordgo.PermissionManageChannels != 0
}
