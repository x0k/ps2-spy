package population_command_handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/bot/render"
	"github.com/x0k/ps2-spy/internal/lib/loaders"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func handleGlobalPopulation(
	ctx context.Context,
	log *logger.Logger,
	opts []*discordgo.ApplicationCommandInteractionDataOption,
	popLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.WorldsPopulation]],
) (*discordgo.WebhookEdit, *handlers.Error) {
	const op = "bot.handlers.command.population_command_handler.handleGlobalPopulation"
	var provider string
	if len(opts) > 0 {
		provider = opts[0].StringValue()
	}
	log.Debug(ctx, "parsed options", slog.String("provider", provider))
	population, err := popLoader.Load(ctx, provider)
	if err != nil {
		return nil, &handlers.Error{
			Msg: "Failed to get population",
			Err: fmt.Errorf("%s error getting population: %w", op, err),
		}
	}
	embeds := []*discordgo.MessageEmbed{
		render.RenderPopulation(population),
	}
	return &discordgo.WebhookEdit{
		Embeds: &embeds,
	}, nil
}

func handleServerPopulation(
	ctx context.Context,
	log *logger.Logger,
	opts []*discordgo.ApplicationCommandInteractionDataOption,
	worldPopLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]],
) (*discordgo.WebhookEdit, *handlers.Error) {
	const op = "bot.handlers.command.population_command_handler.handleServerPopulation"
	server := opts[0].StringValue()
	var provider string
	if len(opts) > 1 {
		provider = opts[1].StringValue()
	}
	log.Debug(ctx, "parsed options", slog.String("server", server), slog.String("provider", provider))
	population, err := worldPopLoader.Load(ctx, loaders.NewMultiLoaderQuery(provider, ps2.WorldId(server)))
	if err != nil {
		return nil, &handlers.Error{
			Msg: "Failed to get population",
			Err: fmt.Errorf("%s error getting population: %w", op, err),
		}
	}
	embeds := []*discordgo.MessageEmbed{
		render.RenderWorldDetailedPopulation(population),
	}
	return &discordgo.WebhookEdit{
		Embeds: &embeds,
	}, nil
}

func New(
	log *logger.Logger,
	popLoader loaders.KeyedLoader[string, loaders.Loaded[ps2.WorldsPopulation]],
	worldPopLoader loaders.QueriedLoader[loaders.MultiLoaderQuery[ps2.WorldId], loaders.Loaded[ps2.DetailedWorldPopulation]],
) handlers.InteractionHandler {
	return handlers.DeferredEphemeralResponse(func(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.WebhookEdit, *handlers.Error) {
		const op = "bot.handlers.command.population_command_handler.New"
		option := i.ApplicationCommandData().Options[0]
		populationType := option.Name // `global` or `server`
		switch populationType {
		case "global":
			return handleGlobalPopulation(ctx, log, option.Options, popLoader)
		case "server":
			return handleServerPopulation(ctx, log, option.Options, worldPopLoader)
		default:
			return nil, &handlers.Error{
				Msg: "Invalid population type",
				Err: fmt.Errorf("%s invalid population type: %s", op, populationType),
			}
		}
	})
}
