package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func serverNames() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(ps2.WorldNames))
	for k, v := range ps2.WorldNames {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: k,
		})
	}
	return choices
}

func providerChoices(providers []string) []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(providers))
	for _, v := range providers {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: v,
		})
	}
	return choices
}

func makeCommands(service *ps2.Service) []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "population",
			Description: "Returns the global population.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "provider",
					Description: "Provider name",
					Choices:     providerChoices(service.PopulationLoaders()),
				},
			},
		},
		{
			Name:        "server-population",
			Description: "Returns the server population.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "server",
					Description: "Server name",
					Choices:     serverNames(),
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "provider",
					Description: "Provider name",
					Choices:     providerChoices(service.PopulationByWorldIdProviders()),
				},
			},
		},
		{
			Name:        "alerts",
			Description: "Returns the alerts.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "server",
					Description: "Server name",
					Choices:     serverNames(),
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "provider",
					Description: "Provider name",
					Choices:     providerChoices(service.AlertsLoaders()),
				},
			},
		},
	}
}
