package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func renderStatsByFactions(p ps2.StatsByFactions) string {
	builder := strings.Builder{}
	builder.Grow(60) // 17 characters per line
	if p.All == 0 {
		builder.WriteString("TR:   0 | 0.00%\nNC:   0 | 0.00%\nVS:   0 | 0.00%\n")
	} else {
		builder.WriteString(fmt.Sprintf("TR: %3d | %.2f%%\n", p.TR, float64(p.TR)/float64(p.All)*100))
		builder.WriteString(fmt.Sprintf("NC: %3d | %.2f%%\n", p.NC, float64(p.NC)/float64(p.All)*100))
		builder.WriteString(fmt.Sprintf("VS: %3d | %.2f%%\n", p.VS, float64(p.VS)/float64(p.All)*100))
		// builder.WriteString(fmt.Sprintf("Other: %3d | %.2f%\n", worldPopulation.Total.Other, float64(worldPopulation.Total.Other)/float64(worldPopulation.Total.All)*100))
	}
	return builder.String()
}

func renderWorldDetailedPopulation(worldPopulation ps2.WorldPopulation, populationSource string, updatedAt time.Time) *discordgo.MessageEmbed {
	zones := make([]*discordgo.MessageEmbedField, 0, len(worldPopulation.Zones))
	for _, zonePopulation := range worldPopulation.Zones {
		if zonePopulation.IsOpen {
			zones = append(zones, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s - %d", zonePopulation.Name, zonePopulation.All),
				Value:  renderStatsByFactions(zonePopulation.StatsByFactions),
				Inline: true,
			})
		}
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.Total.All),
		Fields: zones,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %s", populationSource),
		},
		Timestamp: updatedAt.Format(time.RFC3339),
	}
}

func renderWorldTotalPopulation(worldPopulation ps2.WorldPopulation) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.Total.All),
		Value:  renderStatsByFactions(worldPopulation.Total),
		Inline: true,
	}
}

func renderPopulation(population ps2.WorldsPopulation, populationSource string, updatedAt time.Time) *discordgo.MessageEmbed {
	worlds := make([]ps2.WorldPopulation, 0, len(population.Worlds))
	for _, worldPopulation := range population.Worlds {
		worlds = append(worlds, worldPopulation)
	}
	sort.Slice(worlds, func(i, j int) bool {
		return worlds[i].Id < worlds[j].Id
	})
	fields := make([]*discordgo.MessageEmbedField, len(worlds))
	for i, world := range worlds {
		fields[i] = renderWorldTotalPopulation(world)
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  fmt.Sprintf("Total population - %d", population.Total.All),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %q", populationSource),
		},
		Timestamp: updatedAt.Format(time.RFC3339),
	}
}
