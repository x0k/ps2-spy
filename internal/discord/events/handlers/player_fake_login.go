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
		for _, channel := range e.Channels {
			if !channel.TitleUpdates {
				continue
			}
			updateOnlineCountInTitle(
				ctx,
				m.log,
				session,
				messages,
				channel,
				onlineTrackableEntitiesCountLoader,
				channelTitleUpdater,
			)
		}
		return nil
	})
}
