package discord

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Handler func(
	ctx context.Context,
	session *discordgo.Session,
	channelIds []ChannelId,
	event Event,
) error

type HandlerFactory = func(platform ps2_platforms.Platform) Handler

func SimpleMessage[E Event](handle func(ctx context.Context, e E) (string, *Error)) Handler {
	return func(ctx context.Context, session *discordgo.Session, channelIds []ChannelId, event Event) error {
		const op = "discord.SimpleMessage"
		msg, err := handle(ctx, event.(E))
		if err != nil {
			msg = err.Msg
		}
		if msg == "" {
			return nil
		}
		errs := make([]error, 0, len(channelIds))
		for len(msg) > 0 {
			toSend := msg
			if len(toSend) > 4000 {
				toSend = toSend[:4000]
				lastLineBreak := strings.LastIndexByte(toSend, '\n')
				if lastLineBreak > 0 {
					toSend = toSend[:lastLineBreak]
					msg = msg[lastLineBreak+1:]
				} else {
					lastSpace := strings.LastIndexByte(toSend, ' ')
					if lastSpace > 0 {
						toSend = toSend[:lastSpace]
						msg = msg[lastSpace+1:]
					} else {
						const truncation = "... (truncated)"
						toSend = msg[:4000-len(truncation)] + truncation
						msg = msg[4000-len(truncation):]
					}
				}
			} else {
				msg = ""
			}
			for _, channelId := range channelIds {
				_, err := session.ChannelMessageSend(string(channelId), toSend)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		if err != nil {
			return fmt.Errorf("%s handling event %q: %w", op, event.Type(), err.Err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("%s sending messages: %s", op, errors.Join(errs...))
		}
		return nil
	}
}
