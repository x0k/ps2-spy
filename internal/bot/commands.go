package bot

import (
	"github.com/bwmarrin/discordgo"
	multiloaders "github.com/x0k/ps2-spy/internal/loaders/multi"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
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

func NewCommands(
	popMultiLoader multiloaders.MultiLoader,
	worldPopMultiLoader multiloaders.MultiLoader,
	alertsMultiLoader multiloaders.MultiLoader,
) []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "population",
			Description: "Returns the global population.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "provider",
					Description: "Provider name",
					Choices:     providerChoices(popMultiLoader.Loaders()),
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
					Choices:     providerChoices(worldPopMultiLoader.Loaders()),
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
					Choices:     providerChoices(alertsMultiLoader.Loaders()),
				},
			},
		},
		{
			Name:        "setup",
			Description: "Manage subscription settings for this channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        platforms.PC,
					Description: "Subscription settings for the PC platform",
				},
			},
		},
	}
}
