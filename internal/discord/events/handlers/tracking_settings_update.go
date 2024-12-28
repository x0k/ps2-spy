package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"github.com/x0k/ps2-spy/internal/tracking"
)

type TrackingSettingsDiffViewLoader = func(
	context.Context, discord.ChannelId, ps2_platforms.Platform, tracking.SettingsDiff,
) (tracking.SettingsDiffView, error)

func NewTrackingSettingsUpdateHandler(
	m *HandlersManager,
	messages *discord_messages.Messages,
	trackingSettingsDiffViewLoader TrackingSettingsDiffViewLoader,
) Handler {
	return newHandler(m, func(
		ctx context.Context, s *discordgo.Session, e discord_events.ChannelTrackingSettingsUpdated,
	) error {
		return sendSimpleMessage(
			s, []discord.Channel{e.Channel}, func() discord.Message {
				diffView, err := trackingSettingsDiffViewLoader(ctx, e.Channel.Id, e.Event.Platform, e.Event.Diff)
				if err != nil {
					return discord_messages.TrackingSettingsLoadError[string](
						e.Channel.Id,
						e.Event.Platform,
						err,
					)
				}
				return messages.TrackingSettingsUpdated(
					e.Event.Updater,
					diffView,
				)
			}(),
		)
	})
}
