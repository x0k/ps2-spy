package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewTrackingSettingsUpdateHandler(
	m *HandlersManager,
	messages *discord_messages.Messages,
) Handler {
	return newHandler(m, func(
		ctx context.Context, s *discordgo.Session, e discord_events.ChannelTrackingSettingsUpdated,
	) error {
		return sendSimpleMessage(
			s, []discord.Channel{e.Channel}, messages.TrackingSettingsUpdated(e),
		)
	})
}
