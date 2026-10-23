package discord_event_handlers

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/logger"
)

func updateTitleForChannels(
	ctx context.Context,
	log *logger.Logger,
	session *discordgo.Session,
	messages *discord_messages.Messages,
	channels []discord.Channel,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) {
	for _, channel := range channels {
		if !channel.TitleUpdates {
			continue
		}
		updateOnlineCountInTitle(
			ctx,
			log,
			session,
			messages,
			channel,
			onlineTrackableEntitiesCountLoader,
			channelTitleUpdater,
		)
	}
}
