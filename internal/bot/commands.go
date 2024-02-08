package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/loaders/multi_loaders"
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

func newCommands(
	popMultiLoader multi_loaders.MultiLoader,
	worldPopMultiLoader multi_loaders.MultiLoader,
	alertsMultiLoader multi_loaders.MultiLoader,
) []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "population",
			Description: "Returns the population.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "global",
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
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "server",
					Description: "Returns the server population.",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
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
			},
		},
		{
			Name:        "territories",
			Description: "Returns the server territories control.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "server",
					Description: "Server name",
					Choices:     serverNames(),
					Required:    true,
				},
			},
		},
		{
			Name:        "alerts",
			Description: "Returns the alerts.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
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
					Name:        string(platforms.PC),
					Description: "Subscription settings for the PC platform",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(platforms.PS4_EU),
					Description: "Subscription settings for the PS4 EU platform",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(platforms.PS4_US),
					Description: "Subscription settings for the PS4 US platform",
				},
			},
		},
		{
			Name:        "online",
			Description: "Returns online trackable outfits members and characters",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(platforms.PC),
					Description: "For PC platform",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(platforms.PS4_EU),
					Description: "For PS4 EU platform",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        string(platforms.PS4_US),
					Description: "For PS4 US platform",
				},
			},
		},
		{
			Name:        "about",
			Description: "About this bot",
		},
	}
}
