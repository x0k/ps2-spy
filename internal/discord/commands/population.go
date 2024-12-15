package discord_commands

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func NewPopulation(
	log *logger.Logger,
	messages *discord_messages.Messages,
	populationLoader loader.Keyed[string, meta.Loaded[ps2.WorldsPopulation]],
	populationProviders iter.Seq[string],
	worldPopulationLoader loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]],
	worldPopulationProviders iter.Seq[string],
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "population",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "популяция",
			},
			Description: "Returns the population.",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Возвращает популяцию.",
			},
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionSubCommand,
					Name: "global",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "глобальная",
					},
					Description: "Returns the global population.",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Возвращает глобальную популяцию.",
					},
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type: discordgo.ApplicationCommandOptionString,
							Name: "provider",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "источник",
							},
							Description: "Provider name",
							DescriptionLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "Название провайдера",
							},
							Choices: providerChoices(populationProviders),
						},
					},
				},
				{
					Type: discordgo.ApplicationCommandOptionSubCommand,
					Name: "server",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "сервер",
					},
					Description: "Returns the server population.",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Возвращает популяцию сервера.",
					},
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type: discordgo.ApplicationCommandOptionString,
							Name: "server",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "сервер",
							},
							Description: "Server name",
							DescriptionLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "Название сервера",
							},
							Choices:  serverNames(),
							Required: true,
						},
						{
							Type: discordgo.ApplicationCommandOptionString,
							Name: "provider",
							NameLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "источник",
							},
							Description: "Provider name",
							DescriptionLocalizations: map[discordgo.Locale]string{
								discordgo.Russian: "Название провайдера",
							},
							Choices: providerChoices(worldPopulationProviders),
						},
					},
				},
			},
		},
		Handler: discord.DeferredEphemeralEdit(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.Edit {
			const op = "discord_commands.NewPopulation.Handle"
			option := i.ApplicationCommandData().Options[0]
			populationType := option.Name
			switch populationType {
			case "global":
				return handleGlobalPopulation(ctx, log, messages, option.Options, populationLoader)
			case "server":
				return handleServerPopulation(ctx, log, messages, option.Options, worldPopulationLoader)
			default:
				return messages.InvalidPopulationType(
					populationType,
					fmt.Errorf("%s invalid population type: %s", op, populationType),
				)
			}
		}),
	}
}

func handleGlobalPopulation(
	ctx context.Context,
	log *logger.Logger,
	messages *discord_messages.Messages,
	opts []*discordgo.ApplicationCommandInteractionDataOption,
	popLoader loader.Keyed[string, meta.Loaded[ps2.WorldsPopulation]],
) discord.Edit {
	var provider string
	if len(opts) > 0 {
		provider = opts[0].StringValue()
	}
	log.Debug(ctx, "parsed options", slog.String("provider", provider))
	population, err := popLoader(ctx, provider)
	if err != nil {
		return messages.GlobalPopulationLoadError(provider, err)
	}
	return messages.GlobalPopulation(population)
}

func handleServerPopulation(
	ctx context.Context,
	log *logger.Logger,
	messages *discord_messages.Messages,
	opts []*discordgo.ApplicationCommandInteractionDataOption,
	worldPopLoader loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.DetailedWorldPopulation]],
) discord.Edit {
	server := opts[0].StringValue()
	var provider string
	if len(opts) > 1 {
		provider = opts[1].StringValue()
	}
	log.Debug(ctx, "parsed options", slog.String("server", server), slog.String("provider", provider))
	worldId := ps2.WorldId(server)
	population, err := worldPopLoader(ctx, newQuery(provider, worldId))
	if err != nil {
		return messages.WorldPopulationLoadError(provider, worldId, err)
	}
	return messages.WorldPopulation(population)
}
