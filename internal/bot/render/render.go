package render

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func RenderStatsByFactions(p ps2.StatsByFactions) string {
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

func RenderWorldDetailedPopulation(loaded ps2.Loaded[ps2.DetailedWorldPopulation]) *discordgo.MessageEmbed {
	worldPopulation := loaded.Value
	zones := make([]*discordgo.MessageEmbedField, 0, len(worldPopulation.Zones))
	for _, zonePopulation := range worldPopulation.Zones {
		if zonePopulation.IsOpen {
			zones = append(zones, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s - %d", zonePopulation.Name, zonePopulation.All),
				Value:  RenderStatsByFactions(zonePopulation.StatsByFactions),
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

func RenderWorldTotalPopulation(worldPopulation ps2.WorldPopulation) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.StatsByFactions.All),
		Value:  RenderStatsByFactions(worldPopulation.StatsByFactions),
		Inline: true,
	}
}

func RenderPopulation(loaded ps2.Loaded[ps2.WorldsPopulation]) *discordgo.MessageEmbed {
	population := loaded.Value
	worlds := slices.Clone(population.Worlds)
	sort.Slice(worlds, func(i, j int) bool {
		return worlds[i].Id < worlds[j].Id
	})
	fields := make([]*discordgo.MessageEmbedField, len(worlds))
	for i, world := range worlds {
		fields[i] = RenderWorldTotalPopulation(world)
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

func RenderAlertTitle(alert ps2.Alert) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:  alert.AlertName,
		Value: alert.AlertDescription,
	}
}

func RenderAlertTiming(alert ps2.Alert) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name: "Period",
		Value: fmt.Sprintf(
			"<t:%d:t> - <t:%d:t> (Ends <t:%d:R>)",
			alert.StartedAt.Unix(),
			alert.StartedAt.Add(alert.Duration).Unix(),
			alert.StartedAt.Add(alert.Duration).Unix(),
		),
	}
}

func RenderAlertTerritoryControl(alert ps2.Alert) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:  "Territory Control",
		Value: RenderStatsByFactions(alert.TerritoryControl),
	}
}

func RenderWorldAlerts(alerts ps2.Alerts) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, len(alerts)*3)
	for _, v := range alerts {
		fields = append(fields, RenderAlertTitle(v), RenderAlertTiming(v), RenderAlertTerritoryControl(v))
	}
	return fields
}

func RenderWorldDetailedAlerts(worldName string, loaded ps2.Loaded[ps2.Alerts]) *discordgo.MessageEmbed {
	alerts := loaded.Value
	if len(alerts) == 0 {
		return &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s - No alerts", worldName),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Source: %s", loaded.Source),
			},
			Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
		}
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  fmt.Sprintf("%s alerts", worldName),
		Fields: RenderWorldAlerts(alerts),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %s", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}

func RenderAlerts(loaded ps2.Loaded[ps2.Alerts]) []*discordgo.MessageEmbed {
	alerts := loaded.Value
	if len(alerts) == 0 {
		return []*discordgo.MessageEmbed{
			{
				Title: "No alerts",
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Source: %s", loaded.Source),
				},
				Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
			},
		}
	}
	groups := make(map[ps2.WorldId][]ps2.Alert)
	sortedGroups := make([]ps2.WorldId, 0, len(groups))
	for _, alert := range alerts {
		group, ok := groups[alert.WorldId]
		if !ok {
			sortedGroups = append(sortedGroups, alert.WorldId)
		}
		groups[alert.WorldId] = append(group, alert)
	}
	sort.Slice(sortedGroups, func(i, j int) bool {
		return sortedGroups[i] < sortedGroups[j]
	})
	embeds := make([]*discordgo.MessageEmbed, 0, len(sortedGroups))
	for _, v := range sortedGroups {
		alerts := groups[v]
		worldName := alerts[0].WorldName
		loaded.Value = alerts
		embeds = append(embeds, RenderWorldDetailedAlerts(
			worldName, loaded,
		))
	}
	return embeds
}
