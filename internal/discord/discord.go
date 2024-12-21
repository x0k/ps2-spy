package discord

import (
	"time"

	"golang.org/x/text/language"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/shared"
)

type ChannelId string

type UserId string

type ChannelAndUserIds string

const idsSeparator = "+"

func NewChannelAndUserId(channelId ChannelId, userId UserId) ChannelAndUserIds {
	return ChannelAndUserIds(string(channelId) + idsSeparator + string(userId))
}

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
	Id              StatsTrackerTaskId
	ChannelId       ChannelId
	UtcStartWeekday time.Weekday
	UtcStartTime    time.Duration
	UtcEndWeekday   time.Weekday
	UtcEndTime      time.Duration
}

type StatsTrackerTaskState struct {
	SubmitButtonId string
	TaskId         StatsTrackerTaskId
	Timezone       *time.Location
	LocalWeekdays  []time.Weekday
	LocalStartHour int
	LocalStartMin  int
	Duration       time.Duration
}

func NewCreateStatsTrackerTaskState(
	timezone *time.Location,
) StatsTrackerTaskState {
	localNow := time.Now().In(timezone)
	startTime := time.Duration(localNow.Hour())*time.Hour + time.Duration(localNow.Minute()/10)*10*time.Minute
	hour := int(startTime / time.Hour)
	min := int((startTime % time.Hour) / time.Minute)
	return StatsTrackerTaskState{
		SubmitButtonId: STATS_TRACKER_TASK_CREATE_SUBMIT_BUTTON_CUSTOM_ID,
		Timezone:       timezone,
		LocalWeekdays: []time.Weekday{
			localNow.Weekday(),
		},
		LocalStartHour: hour,
		LocalStartMin:  min,
		Duration:       2 * time.Hour,
	}
}

func NewUpdateStatsTrackerTaskState(
	task StatsTrackerTask,
	timezone *time.Location,
) StatsTrackerTaskState {
	_, offsetInSeconds := time.Now().In(timezone).Zone()
	offset := (time.Duration(offsetInSeconds) * time.Second)
	startWeekday, startTime := shared.ShiftDate(task.UtcStartWeekday, task.UtcStartTime, offset)
	endWeekday, endTime := shared.ShiftDate(task.UtcEndWeekday, task.UtcEndTime, offset)
	duration := endTime - startTime
	if startWeekday != endWeekday {
		duration += 24 * time.Hour
	}
	return StatsTrackerTaskState{
		SubmitButtonId: STATS_TRACKER_TASK_UPDATE_SUBMIT_BUTTON_CUSTOM_ID,
		TaskId:         task.Id,
		Timezone:       timezone,
		LocalWeekdays: []time.Weekday{
			startWeekday,
		},
		LocalStartHour: int(startTime / time.Hour),
		LocalStartMin:  int((startTime % time.Hour) / time.Minute),
		Duration:       duration,
	}
}
