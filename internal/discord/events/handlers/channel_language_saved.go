package discord_event_handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewChannelLanguageSaved(
	m *HandlersManager,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.ChannelLanguageSaved,
	) error {
		updateOnlineCountInTitle(
			ctx,
			m.log.With(slog.String("channel_id", string(e.Channel.Id))),
			session,
			messages,
			e.Channel,
			onlineTrackableEntitiesCountLoader,
			channelTitleUpdater,
		)
		return nil
	})
}
