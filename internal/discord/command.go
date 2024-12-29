package discord

import (
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

const CUSTOM_ID_SEPARATOR = "::"

func HandlerId(customId string) string {
	idx := strings.Index(customId, CUSTOM_ID_SEPARATOR)
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
var TRACKING_EDIT_BUTTON_CUSTOM_ID = "tracking_edit"

func NewTrackingSettingsEditButtonCustomId(
	platform ps2_platforms.Platform,
	outfits []string,
	characters []string,
) string {
	return TRACKING_EDIT_BUTTON_CUSTOM_ID +
		CUSTOM_ID_SEPARATOR + string(platform) +
		CUSTOM_ID_SEPARATOR + strings.Join(outfits, ",") +
		CUSTOM_ID_SEPARATOR + strings.Join(characters, ",")
}

func CustomIdToPlatformAndOutfitsAndCharacters(customId string) (ps2_platforms.Platform, []string, []string) {
	parts := strings.Split(customId, CUSTOM_ID_SEPARATOR)
	return ps2_platforms.Platform(parts[1]), strings.Split(parts[2], ","), strings.Split(parts[3], ",")
}

var CHANNEL_LANGUAGE_COMPONENT_CUSTOM_ID = "channel_language"
var CHANNEL_CHARACTER_NOTIFICATIONS_COMPONENT_CUSTOM_ID = "channel_character_notifications"
var CHANNEL_OUTFIT_NOTIFICATIONS_COMPONENT_CUSTOM_ID = "channel_outfit_notifications"
var CHANNEL_TITLE_UPDATES_COMPONENT_CUSTOM_ID = "channel_title_updates"
var CHANNEL_DEFAULT_TIMEZONE_COMPONENT_CUSTOM_ID = "channel_default_timezone"

var STATS_TRACKER_TASKS_ADD_BUTTON_CUSTOM_ID = "stats_tracker_tasks_add"
var STATS_TRACKER_TASKS_EDIT_BUTTON_CUSTOM_ID = "stats_tracker_tasks_edit"
var STATS_TRACKER_TASKS_REMOVE_BUTTON_CUSTOM_ID = "stats_tracker_tasks_remove"
var STATS_TRACKER_TASKS_PAGE_BUTTON_CUSTOM_ID = "stats_tracker_tasks_page"

var STATS_TRACKER_TASK_WEEKDAYS_SELECTOR_CUSTOM_ID = "stats_tracker_task_weekdays_selector"
var STATS_TRACKER_TASK_START_HOUR_SELECTOR_CUSTOM_ID = "stats_tracker_task_start_time_selector"
var STATS_TRACKER_TASK_START_MINUTE_SELECTOR_CUSTOM_ID = "stats_tracker_task_end_time_selector"
var STATS_TRACKER_TASK_DURATION_SELECTOR_CUSTOM_ID = "stats_tracker_task_duration_selector"
var STATS_TRACKER_TASK_CANCEL_BUTTON_CUSTOM_ID = "stats_tracker_task_cancel"

var STATS_TRACKER_TASK_CREATE_SUBMIT_BUTTON_CUSTOM_ID = "stats_tracker_task_create_submit"
var STATS_TRACKER_TASK_UPDATE_SUBMIT_BUTTON_CUSTOM_ID = "stats_tracker_task_update_submit"
