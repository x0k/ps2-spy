package discord_event_handlers

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_events "github.com/x0k/ps2-spy/internal/discord/events"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CharactersLoader = loader.Multi[ps2.CharacterId, ps2.Character]

func NewOutfitMembersUpdate(
	m *HandlersManager,
	messages *discord_messages.Messages,
	outfitLoader OutfitLoader,
	charactersLoader CharactersLoader,
	platform ps2_platforms.Platform,
) Handler {
	return newHandler(m, func(
		ctx context.Context,
		session *discordgo.Session,
		e discord_events.OutfitMembersUpdate,
	) error {
		return sendSimpleMessage(session, e.Channels, func() discord.Message {
			outfit, err := outfitLoader(ctx, e.Event.OutfitId)
			if err != nil {
				return messages.OutfitLoadError(e.Event.OutfitId, platform, err)
			}
			charIds := append(e.Event.Members.ToAdd, e.Event.Members.ToDel...)
			characters, err := charactersLoader(ctx, charIds)
			if err != nil {
				return messages.CharactersLoadError(charIds, platform, err)
			}
			toAdd := make([]ps2.Character, 0, len(e.Event.Members.ToAdd))
			for _, id := range e.Event.Members.ToAdd {
				char, ok := characters[id]
				if !ok {
					m.log.Warn(ctx, "character not found", slog.String("character_id", string(id)))
					continue
				}
				toAdd = append(toAdd, char)
			}
			toDell := make([]ps2.Character, 0, len(e.Event.Members.ToDel))
			for _, id := range e.Event.Members.ToDel {
				char, ok := characters[id]
				if !ok {
					m.log.Warn(ctx, "character not found", slog.String("character_id", string(id)))
					continue
				}
				toDell = append(toDell, char)
			}
			return messages.OutfitMembersUpdate(outfit, diff.Diff[ps2.Character]{
				ToAdd: toAdd,
				ToDel: toDell,
			})
		}())
	})
}
