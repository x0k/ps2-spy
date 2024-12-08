package discord_event_handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewChannelLanguageUpdate(
	m *HandlersManager,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.ChannelLanguageUpdated,
	) error {
		updateOnlineCountInTitle(
			ctx,
			m.log.With(slog.String("channel_id", string(e.Event.ChannelId))),
			session,
			messages,
			discord.NewChannel(e.Event.ChannelId, e.Event.Language),
			onlineTrackableEntitiesCountLoader,
			channelTitleUpdater,
		)
		return nil
	})
}
