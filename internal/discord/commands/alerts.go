package discord_commands

import (
	"context"
	"iter"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func NewAlerts(
	log *logger.Logger,
	messages discord.LocalizedMessages,
	alertsProviders iter.Seq[string],
	alertsLoader loader.Keyed[string, meta.Loaded[ps2.Alerts]],
	worldAlertsLoader loader.Queried[query[ps2.WorldId], meta.Loaded[ps2.Alerts]],
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "alerts",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "тревоги",
			},
			Description: "Returns the alerts.",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Возвращает тревоги.",
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
					Choices: serverNames(),
				},
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "provider",
					NameLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "провайдер",
					},
					Description: "Provider name",
					DescriptionLocalizations: map[discordgo.Locale]string{
						discordgo.Russian: "Название провайдера",
					},
					Choices: providerChoices(alertsProviders),
				},
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.LocalizedEdit {
			opts := i.ApplicationCommandData().Options
			var worldId ps2.WorldId
			var provider string
			for _, opt := range opts {
				switch opt.Name {
				case "server":
					worldId = ps2.WorldId(opt.StringValue())
				case "provider":
					provider = opt.StringValue()
				}
			}
			log.Debug(ctx, "parsed options", slog.String("world_id", string(worldId)), slog.String("provider", provider))
			if worldId != "" {
				log.Debug(ctx, "getting world alerts")
				alerts, err := worldAlertsLoader(ctx, newQuery(provider, worldId))
				if err != nil {
					return messages.WorldAlertsLoadError(provider, worldId, err)
				}
				worldName := ps2.WorldNameById(worldId)
				return messages.WorldAlerts(worldName, alerts)
			}
			log.Debug(ctx, "getting global alerts")
			alerts, err := alertsLoader(ctx, provider)
			if err != nil {
				return messages.GlobalAlertsLoadError(provider, err)
			}
			return messages.GlobalAlerts(alerts)
		}),
	}
}
