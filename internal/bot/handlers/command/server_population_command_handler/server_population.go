package server_population_command_handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/infra"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func New(
	worldPopLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]],
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		const op = "bot.handlers.command.server_population_command_handler"
		log := infra.OpLogger(ctx, op)
		opts := i.ApplicationCommandData().Options
		server := opts[0].IntValue()
		var provider string
		if len(opts) > 1 {
			provider = opts[1].StringValue()
		}
		log.Debug("parsed options", slog.Int64("server", server), slog.String("provider", provider))
		population, err := worldPopLoader.Load(ctx, loaders.NewMultiLoaderQuery(provider, ps2.WorldId(server)))
		if err != nil {
			return nil, fmt.Errorf("%s error getting population: %q", op, err)
		}
		embeds := []*discordgo.MessageEmbed{
			render.RenderWorldDetailedPopulation(population),
		}
		return &discordgo.WebhookEdit{
			Embeds: &embeds,
		}, nil
	})
}
