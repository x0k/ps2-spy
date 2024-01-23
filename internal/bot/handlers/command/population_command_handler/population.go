package population_command_handler

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
	popLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.WorldsPopulation]],
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		const op = "bot.handlers.command.population_command_handler"
		log := infra.OpLogger(ctx, op)
		opts := i.ApplicationCommandData().Options
		var provider string
		if len(opts) > 0 {
			provider = opts[0].StringValue()
		}
		log.Debug("parsed options", slog.String("provider", provider))
		population, err := popLoader.Load(ctx, provider)
		if err != nil {
			return nil, fmt.Errorf("%s error getting population: %w", op, err)
		}
		embeds := []*discordgo.MessageEmbed{
			render.RenderPopulation(population),
		}
		return &discordgo.WebhookEdit{
			Embeds: &embeds,
		}, nil
	})
}
