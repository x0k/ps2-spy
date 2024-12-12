package discord_event_handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"golang.org/x/text/message"
)

type OnlineTrackableEntitiesCountLoader = loader.Keyed[discord.ChannelId, int]
type ChannelTitleUpdater func(ctx context.Context, channelId discord.ChannelId, title string) error

func updateOnlineCountInTitle(
	ctx context.Context,
	log *logger.Logger,
	session *discordgo.Session,
	messages *discord_messages.Messages,
	channel discord.Channel,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	updateChannelTitle ChannelTitleUpdater,
) {
	trackableEntitiesCount, err := onlineTrackableEntitiesCountLoader(ctx, channel.ChannelId)
	if err != nil {
		log.Error(
			ctx,
			"failed to get trackable entities while updating online count in title",
			slog.String("channel_id", string(channel.ChannelId)),
			sl.Err(err),
		)
		return
	}
	c, err := session.Channel(string(channel.ChannelId))
	if err != nil {
		log.Error(
			ctx,
			"failed to get channel info while updating online count in title",
			slog.String("channel_id", string(channel.ChannelId)),
			sl.Err(err),
		)
		return
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
		return
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
