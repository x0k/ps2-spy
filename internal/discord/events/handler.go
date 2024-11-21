package discord_events

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/logger/sl"
	"github.com/x0k/ps2-spy/internal/lib/slicesx"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Handler func(
	ctx context.Context,
	session *discordgo.Session,
	channelIds []discord.Channel,
	event Event,
) error

type HandlerFactory = func(platform ps2_platforms.Platform) Handler

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

func SimpleMessage[E Event](handle func(ctx context.Context, e E) discord.Message) Handler {
	return func(ctx context.Context, session *discordgo.Session, channels []discord.Channel, event Event) error {
		const op = "discord.SimpleMessage"
		msgRenderer := handle(ctx, event.(E))
		channelsByLocale := slicesx.GroupBy(channels, func(c discord.Channel) language.Tag { return c.Locale })
		handlingErrors := make([]error, 0, len(channels))
		sendErrors := make([]error, 0, len(channels))
		for locale, channels := range channelsByLocale {
			msgStr, err := msgRenderer(message.NewPrinter(locale))
			if err != nil {
				msgStr = err.Msg
				handlingErrors = append(handlingErrors, err.Err)
			}
			msg := []rune(msgStr)
			for len(msg) > 0 {
				toSend := msg
				if len(toSend) > msgMaxLen {
					toSend = toSend[:msgMaxLen]
					lastLineBreak := lastIndexRune(toSend, '\n')
					if lastLineBreak > 0 {
						toSend = toSend[:lastLineBreak]
						msg = msg[lastLineBreak+1:]
					} else {
						lastSpace := lastIndexRune(toSend, ' ')
						if lastSpace > 0 {
							toSend = toSend[:lastSpace]
							msg = msg[lastSpace+1:]
						} else {
							toSend = slices.Concat(msg[:msgMaxLen-len(truncation)], truncation)
							msg = msg[msgMaxLen-len(truncation):]
						}
					}
				} else {
					msg = msg[len(toSend):]
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
			return fmt.Errorf("%s handling event %q: %w", op, event.Type(), errors.Join(handlingErrors...))
		}
		if len(sendErrors) > 0 {
			return fmt.Errorf("%s sending messages: %s", op, errors.Join(sendErrors...))
		}
		return nil
	}
}

func ShowModal(handle func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) discord.Response) discord.InteractionHandler {
	return func(ctx context.Context, log *logger.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) error {
		data, customErr := handle(ctx, s, i)(message.NewPrinter(discord.LangTagFromInteraction(i)))
		if customErr != nil {
			if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: customErr.Msg,
			}); err != nil {
				log.Error(ctx, "error sending followup message", sl.Err(err))
			}
			return customErr.Err
		}
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: data,
		})
	}
}
