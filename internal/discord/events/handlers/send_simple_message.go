package discord_event_handlers

import (
	"errors"
	"fmt"
	"slices"

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

var truncation = []rune("... (truncated)")

const msgMaxLen = 2000

func sendSimpleMessage(session *discordgo.Session, channels []discord.Channel, discordMsg discord.Message) error {
	const op = "discord.SimpleMessage"
	channelsByLocale := slicesx.GroupBy(channels, func(c discord.Channel) language.Tag { return c.Locale })
	handlingErrors := make([]error, 0, len(channels))
	sendErrors := make([]error, 0, len(channels))
	for locale, channels := range channelsByLocale {
		msgContent, err := discordMsg(message.NewPrinter(locale))
		if err != nil {
			msgContent = err.Msg
			handlingErrors = append(handlingErrors, err.Err)
		}
		msgRunes := []rune(msgContent)
		for len(msgRunes) > 0 {
			toSend := msgRunes
			if len(toSend) > msgMaxLen {
				toSend = toSend[:msgMaxLen]
				lastLineBreak := lastIndexRune(toSend, '\n')
				if lastLineBreak > 0 {
					toSend = toSend[:lastLineBreak]
					msgRunes = msgRunes[lastLineBreak+1:]
				} else {
					lastSpace := lastIndexRune(toSend, ' ')
					if lastSpace > 0 {
						toSend = toSend[:lastSpace]
						msgRunes = msgRunes[lastSpace+1:]
					} else {
						toSend = slices.Concat(msgRunes[:msgMaxLen-len(truncation)], truncation)
						msgRunes = msgRunes[msgMaxLen-len(truncation):]
					}
				}
			} else {
				msgRunes = msgRunes[len(toSend):]
			}
			for _, channel := range channels {
				_, err := session.ChannelMessageSend(string(channel.ChannelId), string(toSend))
				if err != nil {
					sendErrors = append(sendErrors, err)
				}
			}
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
