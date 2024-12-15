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
	return newHandler(m, func(ctx context.Context, session *discordgo.Session, event discord_events.ChannelTrackerStopped) error {
		platforms := event.Event.Platforms
		// Unreachable
		if len(platforms) == 0 {
			return nil
		}
		var errs []error
		channels := []discord.Channel{discord.NewChannel(event.Event.ChannelId, event.Language)}
		for platform, stats := range platforms {
			if err := sendChunkableMessage(
				session,
				channels,
				messages.ChannelTrackerStopped(
					platform,
					event.Event.StartedAt,
					event.Event.StoppedAt,
					stats,
				),
			); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	})
}
