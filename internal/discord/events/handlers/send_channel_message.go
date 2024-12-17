package discord_event_handlers

import (
	"errors"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

var truncation = []rune("... (truncated)")

const msgMaxLen = 2000

func sendChannelMessage(
	session *discordgo.Session,
	channels []discord.Channel,
	content string,
) error {
	sendErrors := make([]error, 0, len(channels))
	msgRunes := []rune(content)
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
			_, err := session.ChannelMessageSend(string(channel.Id), string(toSend))
			if err != nil {
				sendErrors = append(sendErrors, err)
			}
		}
	}
	return errors.Join(sendErrors...)
}
