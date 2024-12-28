package discord_event_handlers

import (
	"errors"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/iterx"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func lastIndexRune(runes []rune, target rune) int {
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == target {
			return i
		}
	}
	return -1
}

func sendSimpleMessage(session *discordgo.Session, channels []discord.Channel, discordMsg discord.Message) error {
	if len(channels) == 0 {
		return nil
	}
	channelsByLocale := iterx.GroupBy(
		slices.Values(channels),
		func(c discord.Channel) language.Tag { return c.Locale },
	)
	errs := make([]error, 0, len(channels))
	for locale, channels := range channelsByLocale {
		msgContent, err := discordMsg(message.NewPrinter(locale))
		if err != nil {
			msgContent = err.Msg
			errs = append(errs, err.Err)
		}
		if err := sendChannelMessage(session, channels, msgContent); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("send simple message: %w", errors.Join(errs...))
	}
	return nil
}
