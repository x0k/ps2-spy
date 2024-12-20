package discord

import (
	"time"

	"golang.org/x/text/language"

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

type TrackingSettings = TrackableEntities[[]ps2.OutfitId, []ps2.CharacterId]

func CalculateTrackingSettingsDiff(
	old TrackingSettings,
	new TrackingSettings,
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

var DEFAULT_LANG_TAG = language.English

func LangTagFromInteraction(i *discordgo.InteractionCreate) language.Tag {
	if t, err := language.Parse(string(i.Locale)); err == nil {
		return t
	}
	return DEFAULT_LANG_TAG
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

type StatsTrackerTaskId int64

type StatsTrackerTask struct {
	Id           StatsTrackerTaskId
	ChannelId    ChannelId
	UtcWeekday   time.Weekday
	UtcStartTime time.Duration
	UtcEndTime   time.Duration
}

type CreateStatsTrackerTaskState struct {
	Timezone       *time.Location
	LocalWeekdays  []time.Weekday
	LocalStartTime time.Duration
	LocalEndTime   time.Duration
}

func NewCreateStatsTrackerTaskState(
	timezone *time.Location,
) CreateStatsTrackerTaskState {
	localNow := time.Now().In(timezone)
	startTime := time.Duration(localNow.Hour())*time.Hour + time.Duration(localNow.Minute()/10)*10*time.Minute
	return CreateStatsTrackerTaskState{
		Timezone: timezone,
		LocalWeekdays: []time.Weekday{
			localNow.Weekday(),
		},
		LocalStartTime: startTime,
		LocalEndTime:   startTime + 2*time.Hour,
	}
}
