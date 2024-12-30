package discord_event_handlers

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewChannelTrackerStopped(
	m *HandlersManager,
	messages *discord_messages.Messages,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.ChannelTrackerStopped,
	) error {
		platforms := e.Event.Platforms
		// Unreachable
		if len(platforms) == 0 {
			return nil
		}
		var errs []error
		channels := []discord.Channel{e.Channel}
		for _, stats := range platforms {
			if err := sendChunkableMessage(
				session,
				channels,
				messages.ChannelTrackerStopped(
					stats,
					e.Event.StartedAt,
					e.Event.StoppedAt,
				),
			); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	})
}
