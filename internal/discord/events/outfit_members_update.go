package discord_events

import (
	"context"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func NewOutfitMembersUpdateHandlerFactory(
	log *logger.Logger,
	messages *discord_messages.Messages,
	outfitLoaders map[ps2_platforms.Platform]loader.Keyed[ps2.OutfitId, ps2.Outfit],
	charactersLoaders map[ps2_platforms.Platform]loader.Multi[ps2.CharacterId, ps2.Character],
) HandlerFactory {
	return func(platform ps2_platforms.Platform) Handler {
		outfitLoader := outfitLoaders[platform]
		charsLoader := charactersLoaders[platform]
		return SimpleMessage(func(ctx context.Context, event OutfitMembersUpdate) discord.Message {
			outfit, err := outfitLoader(ctx, event.OutfitId)
			if err != nil {
				return messages.OutfitLoadError(event.OutfitId, platform, err)
			}
			charIds := append(event.Members.ToAdd, event.Members.ToDel...)
			characters, err := charsLoader(ctx, charIds)
			if err != nil {
				return messages.CharactersLoadError(charIds, platform, err)
			}
			toAdd := make([]ps2.Character, 0, len(event.Members.ToAdd))
			for _, id := range event.Members.ToAdd {
				char, ok := characters[id]
				if !ok {
					log.Warn(ctx, "character not found", slog.String("character_id", string(id)))
					continue
				}
				toAdd = append(toAdd, char)
			}
			toDell := make([]ps2.Character, 0, len(event.Members.ToDel))
			for _, id := range event.Members.ToDel {
				char, ok := characters[id]
				if !ok {
					log.Warn(ctx, "character not found", slog.String("character_id", string(id)))
					continue
				}
				toDell = append(toDell, char)
			}
			return messages.OutfitMembersUpdate(outfit, diff.Diff[ps2.Character]{
				ToAdd: toAdd,
				ToDel: toDell,
			})
		})
	}
}
