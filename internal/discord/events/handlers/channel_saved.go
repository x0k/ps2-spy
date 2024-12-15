package discord_event_handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewChannelSaved(
	m *HandlersManager,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.ChannelSaved,
	) error {
		if e.Event.OldChannel.TitleUpdates == e.Event.NewChannel.TitleUpdates {
			return nil
		}
		updateOnlineCountInTitle(
			ctx,
			m.log.With(slog.String("channel_id", string(e.Event.NewChannel.Id))),
			session,
			messages,
			e.Event.NewChannel,
			onlineTrackableEntitiesCountLoader,
			channelTitleUpdater,
		)
		return nil
	})
}
