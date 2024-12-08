package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CharacterLoader = loader.Keyed[ps2.CharacterId, ps2.Character]

func NewPlayerLogin(
	m *HandlersManager,
	messages *discord_messages.Messages,
	characterLoader CharacterLoader,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(ctx context.Context, session *discordgo.Session, event discord_events.PlayerLogin) error {
		for _, channel := range event.Channels {
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
		return sendSimpleMessage(session, event.Channels, func() discord.Message {
			char, err := characterLoader(ctx, event.Event.CharacterId)
			if err != nil {
				return messages.CharacterLoadError(event.Event.CharacterId, err)
			}
			return messages.CharacterLogin(char)
		}())
	})
}
