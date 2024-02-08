package territories_command_handler

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	worldTerritoryControlLoader loaders.KeyedLoader[ps2.WorldId, loaders.Loaded[ps2.WorldTerritoryControl]],
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(
		ctx context.Context,
		s *discordgo.Session,
		i *discordgo.InteractionCreate,
	) (*discordgo.WebhookEdit, *handlers.Error) {
		server := i.ApplicationCommandData().Options[0].StringValue()
		loaded, err := worldTerritoryControlLoader.Load(ctx, ps2.WorldId(server))
		if err != nil {
			return nil, &handlers.Error{
				Msg: "Failed to get territory control data",
				Err: err,
			}
		}
		embeds := []*discordgo.MessageEmbed{
			render.RenderWorldTerritoriesControl(loaded),
		}
		return &discordgo.WebhookEdit{
			Embeds: &embeds,
		}, nil
	})
}
