package discord_event_handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/logger"
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
		return channelSavedHandlerBody(ctx, m.log, session, messages, e.Channel,
			onlineTrackableEntitiesCountLoader, channelTitleUpdater)
	})
}

func NewChannelTitleUpdatesSaved(
	m *HandlersManager,
	messages *discord_messages.Messages,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.ChannelTitleUpdatesSaved,
	) error {
		return channelSavedHandlerBody(ctx, m.log, session, messages, e.Channel,
			onlineTrackableEntitiesCountLoader, channelTitleUpdater)
	})
}

func channelSavedHandlerBody(
	ctx context.Context,
	log *logger.Logger,
	session *discordgo.Session,
	messages *discord_messages.Messages,
	channel discord.Channel,
	onlineTrackableEntitiesCountLoader OnlineTrackableEntitiesCountLoader,
	channelTitleUpdater ChannelTitleUpdater,
) error {
	updateOnlineCountInTitle(
		ctx,
		log.With(slog.String("channel_id", string(channel.Id))),
		session,
		messages,
		channel,
		onlineTrackableEntitiesCountLoader,
		channelTitleUpdater,
	)
	return nil
}
