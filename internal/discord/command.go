package discord

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Command struct {
	Cmd               *discordgo.ApplicationCommand
	Handler           InteractionHandler
	SubmitHandlers    map[string]InteractionHandler
	ComponentHandlers map[string]InteractionHandler
}

const customIdSeparator = "::"

func HandlerId(customId string) string {
	idx := strings.Index(customId, customIdSeparator)
	if idx == -1 {
		return customId
	}
	return customId[:idx]
}

var TRACKING_MODAL_CUSTOM_IDS = map[ps2_platforms.Platform]string{
	ps2_platforms.PC:     "tracking_setup_pc",
	ps2_platforms.PS4_EU: "tracking_setup_ps4_eu",
	ps2_platforms.PS4_US: "tracking_setup_ps4_us",
}

var CHANNEL_LANGUAGE_COMPONENT_CUSTOM_ID = "channel_language"
var CHANNEL_CHARACTER_NOTIFICATIONS_COMPONENT_CUSTOM_ID = "channel_character_notifications"
var CHANNEL_OUTFIT_NOTIFICATIONS_COMPONENT_CUSTOM_ID = "channel_outfit_notifications"
var CHANNEL_TITLE_UPDATES_COMPONENT_CUSTOM_ID = "channel_title_updates"
var CHANNEL_DEFAULT_TIMEZONE_COMPONENT_CUSTOM_ID = "channel_default_timezone"

var STATS_TRACKER_TASK_ADD_BUTTON_CUSTOM_ID = "stats_tracker_task_add"
var STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID = "stats_tracker_task_edit"
var STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID = "stats_tracker_task_remove"
var STATS_TRACKER_CREATE_TASK_WEEKDAYS_SELECTOR_CUSTOM_ID = "stats_tracker_task_weekdays_selector"
var STATS_TRACKER_CREATE_TASK_START_HOUR_SELECTOR_CUSTOM_ID = "stats_tracker_task_start_time_selector"
var STATS_TRACKER_CREATE_TASK_START_MINUTE_SELECTOR_CUSTOM_ID = "stats_tracker_task_end_time_selector"
var STATS_TRACKER_CREATE_TASK_DURATION_SELECTOR_CUSTOM_ID = "stats_tracker_task_duration_selector"
var STATS_TRACKER_CREATE_TASK_SUBMIT_BUTTON_CUSTOM_ID = "stats_tracker_task_submit"
var STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID = "stats_tracker_task_page_button"
var STATS_TRACKER_CREATE_TASK_CANCEL_BUTTON_CUSTOM_ID = "stats_tracker_create_task_cancel"

func NewStatsTrackerTaskPageButtonCustomId(
	page int,
) string {
	return STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID + customIdSeparator +
		strconv.Itoa(page)
}

func CustomIdToPage(customId string) (int, error) {
	return strconv.Atoi(
		customId[len(STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID)+len(customIdSeparator):],
	)
}

func NewStatsTrackerTaskEditButtonCustomId(
	id StatsTrackerTaskId,
) string {
	return STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID + customIdSeparator +
		strconv.FormatInt(int64(id), 10)
}

func CustomIdToTaskIdToEdit(customId string) (StatsTrackerTaskId, error) {
	v, err := strconv.ParseInt(
		customId[len(STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID)+len(customIdSeparator):],
		10,
		64,
	)
	return StatsTrackerTaskId(v), err
}

func NewStatsTrackerTaskRemoveButtonCustomId(
	id StatsTrackerTaskId,
) string {
	return STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID + customIdSeparator +
		strconv.FormatInt(int64(id), 10)
}

func CustomIdToTaskIdToRemove(customId string) (StatsTrackerTaskId, error) {
	v, err := strconv.ParseInt(
		customId[len(STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID)+len(customIdSeparator):],
		10,
		64,
	)
	return StatsTrackerTaskId(v), err
}
