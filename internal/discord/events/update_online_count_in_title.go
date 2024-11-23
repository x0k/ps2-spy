package discord_events

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/message"
)

type ChannelTitleUpdater func(ctx context.Context, channelId discord.ChannelId, title string) error

func NewUpdateOnlineCountInTitleHandlerFactory(
	log *logger.Logger,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader loader.Keyed[discord.ChannelId, int],
	updateChannelTitle ChannelTitleUpdater,
) HandlerFactory {
	return func(platform ps2_platforms.Platform) Handler {
		return func(ctx context.Context, session *discordgo.Session, channelIds []discord.Channel, event Event) error {
			for _, channel := range channelIds {
				trackableEntitiesCount, err := onlineTrackableEntitiesCountLoader(ctx, channel.ChannelId)
				if err != nil {
					log.Error(
						ctx,
						"failed to get trackable entities while updating online count in title",
						slog.String("channel_id", string(channel.ChannelId)),
						sl.Err(err),
					)
					continue
				}
				c, err := session.Channel(string(channel.ChannelId))
				if err != nil {
					log.Error(
						ctx,
						"failed to get channel info while updating online count in title",
						slog.String("channel_id", string(channel.ChannelId)),
						sl.Err(err),
					)
					continue
				}
				newTitle, err2 := messages.OnlineCountTitleUpdate(c.Name, trackableEntitiesCount)(
					message.NewPrinter(channel.Locale),
				)
				if err2 != nil {
					log.Error(
						ctx,
						"failed to get new title while update online count in title",
						slog.String("channel_id", string(channel.ChannelId)),
						sl.Err(err2.Err),
					)
					continue
				}
				if err := updateChannelTitle(ctx, channel.ChannelId, newTitle); err != nil {
					log.Error(
						ctx,
						"failed to update online count in title",
						slog.String("channel_id", string(channel.ChannelId)),
						sl.Err(err),
					)
				}
			}
			return nil
		}
	}
}
