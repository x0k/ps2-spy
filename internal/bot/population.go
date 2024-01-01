package bot

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func renderStatsByFactions(p ps2.StatsByFactions) string {
	builder := strings.Builder{}
	builder.Grow(60) // 16 characters per line
	if p.All == 0 {
		builder.WriteString("TR:   0 | 0.0%\nNC:   0 | 0.0%\nVS:   0 | 0.0%\n")
	} else {
		builder.WriteString(fmt.Sprintf("TR: %3d | %.1f%%\n", p.TR, float64(p.TR)/float64(p.All)*100))
		builder.WriteString(fmt.Sprintf("NC: %3d | %.1f%%\n", p.NC, float64(p.NC)/float64(p.All)*100))
		builder.WriteString(fmt.Sprintf("VS: %3d | %.1f%%\n", p.VS, float64(p.VS)/float64(p.All)*100))
		// builder.WriteString(fmt.Sprintf("Other: %3d | %.2f%\n", worldPopulation.Total.Other, float64(worldPopulation.Total.Other)/float64(worldPopulation.Total.All)*100))
	}
	return builder.String()
}

func renderWorldDetailedPopulation(loaded ps2.Loaded[ps2.DetailedWorldPopulation]) *discordgo.MessageEmbed {
	worldPopulation := loaded.Value
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
		Title:  fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.Total),
		Fields: zones,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %s", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}

func renderWorldTotalPopulation(worldPopulation ps2.WorldPopulation) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.StatsByFactions.All),
		Value:  renderStatsByFactions(worldPopulation.StatsByFactions),
		Inline: true,
	}
}

func renderPopulation(loaded ps2.Loaded[ps2.WorldsPopulation]) *discordgo.MessageEmbed {
	population := loaded.Value
	worlds := slices.Clone(population.Worlds)
	sort.Slice(worlds, func(i, j int) bool {
		return worlds[i].Id < worlds[j].Id
	})
	fields := make([]*discordgo.MessageEmbedField, len(worlds))
	for i, world := range worlds {
		fields[i] = renderWorldTotalPopulation(world)
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  fmt.Sprintf("Total population - %d", population.Total),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %q", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}
