package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type CharacterLoader = loader.Keyed[ps2.CharacterId, ps2.Character]

func NewPlayerLogout(
	m *HandlersManager,
	messages *discord_messages.Messages,
	characterLoader CharacterLoader,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(ctx context.Context, session *discordgo.Session, e discord_events.PlayerLogout) error {
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
		return sendSimpleMessage(
			session,
			slicesx.Filter(e.Channels, func(i int) bool {
				return e.Channels[i].CharacterNotifications
			}),
			func() discord.Message {
				char, err := characterLoader(ctx, e.Event.CharacterId)
				if err != nil {
					return messages.CharacterLoadError(e.Event.CharacterId, err)
				}
				return messages.CharacterLogout(char)
			}(),
		)
	})
}
