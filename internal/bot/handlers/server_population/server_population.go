package serverpopulation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type WorldPopulationProvider interface {
	PopulationByWorldId(ctx context.Context, worldId ps2.WorldId, provider string) (ps2.Loaded[ps2.DetailedWorldPopulation], error)
}

func New(worldPopProvider WorldPopulationProvider) handlers.InteractionHandler {
	return handlers.DeferredResponse(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, error) {
		const op = "handlers.server_population"
		log = log.With(slog.String("op", op))
		opts := i.ApplicationCommandData().Options
		log.Debug("command options", slog.Any("options", opts))
		server := opts[0].IntValue()
		var provider string
		if len(opts) > 1 {
			provider = opts[1].StringValue()
		}
		log.Debug("parsed options", slog.Int64("server", server), slog.String("provider", provider))
		population, err := worldPopProvider.PopulationByWorldId(ctx, ps2.WorldId(server), provider)
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
