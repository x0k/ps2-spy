package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
)

func NewPlayerFakeLogin(
	m *HandlersManager,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.PlayerFakeLogin,
	) error {
		updateTitleForChannels(ctx, m.log, session, messages, e.Channels,
			onlineTrackableEntitiesCountLoader, channelTitleUpdater)
		return nil
	})
}
