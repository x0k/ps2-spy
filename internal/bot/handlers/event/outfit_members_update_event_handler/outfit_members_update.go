package outfit_members_update_event_handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/savers/outfit_members_saver"
)

func New(
	log *logger.Logger,
	outfitLoader loaders.KeyedLoader[ps2.OutfitId, ps2.Outfit],
	charsLoader loaders.QueriedLoader[[]ps2.CharacterId, map[ps2.CharacterId]ps2.Character],
) handlers.Ps2EventHandler[outfit_members_saver.OutfitMembersUpdate] {
	return handlers.SimpleMessage[outfit_members_saver.OutfitMembersUpdate](func(
		ctx context.Context,
		cfg *handlers.Ps2EventHandlerConfig,
		event outfit_members_saver.OutfitMembersUpdate,
	) (string, *handlers.Error) {
		const op = "bot.handlers.events.outfit_members_update_event_handler"
		outfit, err := outfitLoader.Load(ctx, event.OutfitId)
		if err != nil {
			return "", &handlers.Error{
				Msg: "Failed to get outfit data",
				Err: fmt.Errorf("%s error getting outfit: %w", op, err),
			}
		}
		characters, err := charsLoader.Load(ctx, append(event.Members.ToAdd, event.Members.ToDel...))
		if err != nil {
			return "", &handlers.Error{
				Msg: "Failed to get character data",
				Err: fmt.Errorf("%s error getting characters: %w", op, err),
			}
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
		return render.RenderOutfitMembersUpdate(outfit, diff.Diff[ps2.Character]{
			ToAdd: toAdd,
			ToDel: toDell,
		}), nil
	})
}
