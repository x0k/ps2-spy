package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewChannelTrackerStarted(
	m *HandlersManager,
	messages *discord_messages.Messages,
) Handler {
	return newHandler(m, func(ctx context.Context, session *discordgo.Session, event discord_events.ChannelTrackerStarted) error {
		return sendSimpleMessage(
			session,
			[]discord.Channel{discord.NewChannel(event.Event.ChannelId, event.Language)},
			messages.ChannelTrackerStarted(),
		)
	})
}
