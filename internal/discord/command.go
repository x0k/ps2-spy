package discord

import (
	"github.com/bwmarrin/discordgo"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Command struct {
	Cmd               *discordgo.ApplicationCommand
	Handler           InteractionHandler
	SubmitHandlers    map[string]InteractionHandler
	ComponentHandlers map[string]InteractionHandler
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
