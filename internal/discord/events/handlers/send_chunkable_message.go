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

func sendChunkableMessage(
	session *discordgo.Session,
	channels []discord.Channel,
	discordMsg discord.ChunkableMessage,
) error {
	if len(channels) == 0 {
		return nil
	}
	channelsByLocale := iterx.GroupBy(
		slices.Values(channels),
		func(c discord.Channel) language.Tag { return c.Locale },
	)
	errs := make([]error, 0, len(channels))
	for locale, channels := range channelsByLocale {
		msg := discordMsg(message.NewPrinter(locale))
		if msg.Chunks == 0 {
			msgContent, err := msg.Print(0, 0)
			if err != nil {
				msgContent = err.Msg
				errs = append(errs, err.Err)
			}
			if err := sendChannelMessage(session, channels, msgContent); err != nil {
				errs = append(errs, err)
			}
			continue
		}
		l := 0
		for l < msg.Chunks {
			r := binSearch(msg.Chunks-l, func(i int) bool {
				content, err := msg.Print(l, i+1)
				if err != nil {
					return false
				}
				return len(content) < msgMaxLen
			})
			count := max(r, 0) + 1
			content, err := msg.Print(l, count)
			if err != nil {
				content = err.Msg
				errs = append(errs, err.Err)
			}
			if err := sendChannelMessage(session, channels, content); err != nil {
				errs = append(errs, err)
			}
			l += count
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("send chunkable message: %w", errors.Join(errs...))
	}
	return nil
}

func binSearch(
	count int,
	probe func(int) bool,
) int {
	if count < 1 {
		return -1
	}
	low := 0
	high := int(count) - 1
	for low < high {
		mid := (low + high) / 2
		if probe(mid) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if !probe(low) {
		low--
	}
	return low
}
