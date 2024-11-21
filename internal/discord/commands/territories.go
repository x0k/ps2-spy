package discord_commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	discord_messages "github.com/x0k/ps2-spy/internal/discord/messages"
	"github.com/x0k/ps2-spy/internal/lib/loader"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func NewTerritories(
	messages *discord_messages.Messages,
	WorldTerritoryControlLoader loader.Keyed[ps2.WorldId, meta.Loaded[ps2.WorldTerritoryControl]],
) *discord.Command {
	return &discord.Command{
		Cmd: &discordgo.ApplicationCommand{
			Name: "territories",
			NameLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "территории",
			},
			Description: "Returns the server territories control.",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.Russian: "Возвращает контролируемые территории сервера.",
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
			},
		},
		Handler: discord.DeferredEphemeralResponse(func(
			ctx context.Context,
			s *discordgo.Session,
			i *discordgo.InteractionCreate,
		) discord.Edit {
			server := i.ApplicationCommandData().Options[0].StringValue()
			worldId := ps2.WorldId(server)
			loaded, err := WorldTerritoryControlLoader(ctx, worldId)
			if err != nil {
				return messages.WorldTerritoryControlLoadError(worldId, err)
			}
			return messages.WorldTerritoryControl(loaded)
		}),
	}
}
