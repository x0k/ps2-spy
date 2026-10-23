package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/db"
	"golang.org/x/text/language"
)

func ChannelFromDTO(ctx context.Context, log *slog.Logger, dto db.Channel) discord.Channel {
	channelId := discord.ChannelId(dto.ChannelID)
	locale, err := language.Parse(dto.Locale)
	if err != nil {
		log.WarnContext(ctx, "failed to parse locale", slog.String("channel_id", string(channelId)), slog.String("locale", dto.Locale), slog.Any("error", err))
		locale = discord.DEFAULT_LANG_TAG
	}
	loc, err := time.LoadLocation(dto.DefaultTimezone)
	if err != nil {
		log.WarnContext(ctx, "failed to load timezone", slog.String("channel_id", string(channelId)), slog.String("timezone", dto.DefaultTimezone), slog.Any("error", err))
		loc = time.UTC
	}
	return discord.NewChannel(
		channelId,
		locale,
		dto.CharacterNotifications,
		dto.OutfitNotifications,
		dto.TitleUpdates,
		loc,
	)
}
