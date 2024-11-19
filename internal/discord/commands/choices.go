package discord_commands

import (
	"iter"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/lib/iterx"
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

func providerChoices(providers iter.Seq[string]) []*discordgo.ApplicationCommandOptionChoice {
	return slices.Collect(iterx.Map(providers, func(v string) *discordgo.ApplicationCommandOptionChoice {
		return &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: v,
		}
	}))
}
