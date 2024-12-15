package discord_event_handlers

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
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
	const op = "discord.sendSimpleMessage"
	channelsByLocale := slicesx.GroupBy(channels, func(c discord.Channel) language.Tag { return c.Locale })
	handlingErrors := make([]error, 0, len(channels))
	sendErrors := make([]error, 0, len(channels))
	for locale, channels := range channelsByLocale {
		msgContent, err := discordMsg(message.NewPrinter(locale))
		if err != nil {
			msgContent = err.Msg
			handlingErrors = append(handlingErrors, err.Err)
		}
		if err := sendChannelMessage(session, channels, msgContent); err != nil {
			sendErrors = append(sendErrors, err)
		}
	}
	if len(handlingErrors) > 0 {
		return fmt.Errorf("%s handling event: %w", op, errors.Join(handlingErrors...))
	}
	if len(sendErrors) > 0 {
		return fmt.Errorf("%s sending messages: %s", op, errors.Join(sendErrors...))
	}
	return nil
}
