package discord_messages

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	"golang.org/x/text/message"
)

func renderStatPerFactions(p *message.Printer, builder *strings.Builder, st ps2.StatPerFactions) {
	builder.Grow(60) // 16 characters per line
	tr := p.Sprintf("TR")
	nc := p.Sprintf("NC")
	vs := p.Sprintf("VS")
	if st.All == 0 {
		builder.WriteString(fmt.Sprintf("%s:   0 | 0.0%%\n%s:   0 | 0.0%%\n%s:   0 | 0.0%%\n", tr, nc, vs))
	} else {
		builder.WriteString(fmt.Sprintf("%s: %3d | %.1f%%\n", tr, st.TR, float64(st.TR)/float64(st.All)*100))
		builder.WriteString(fmt.Sprintf("%s: %3d | %.1f%%\n", nc, st.NC, float64(st.NC)/float64(st.All)*100))
		builder.WriteString(fmt.Sprintf("%s: %3d | %.1f%%\n", vs, st.VS, float64(st.VS)/float64(st.All)*100))
		// builder.WriteString(p.Sprintf("Other: %3d | %.2f%\n", worldPopulation.Total.Other, float64(worldPopulation.Total.Other)/float64(worldPopulation.Total.All)*100))
	}
}

func renderWorldDetailedPopulation(p *message.Printer, loaded meta.Loaded[ps2.DetailedWorldPopulation]) *discordgo.MessageEmbed {
	worldPopulation := loaded.Value
	zones := make([]*discordgo.MessageEmbedField, 0, len(worldPopulation.Zones))
	b := strings.Builder{}
	for _, zonePopulation := range worldPopulation.Zones {
		if zonePopulation.IsOpen {
			renderStatPerFactions(p, &b, zonePopulation.StatPerFactions)
			zones = append(zones, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%s - %d", zonePopulation.Name, zonePopulation.All),
				Value:  b.String(),
				Inline: true,
			})
			b.Reset()
		}
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  fmt.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.Total),
		Fields: zones,
		Footer: &discordgo.MessageEmbedFooter{
			Text: p.Sprintf("Source: %s", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}

func RenderWorldTotalPopulation(p *message.Printer, worldPopulation ps2.WorldPopulation) *discordgo.MessageEmbedField {
	b := strings.Builder{}
	renderStatPerFactions(p, &b, worldPopulation.StatPerFactions)
	return &discordgo.MessageEmbedField{
		Name:   p.Sprintf("%s - %d", worldPopulation.Name, worldPopulation.StatPerFactions.All),
		Value:  b.String(),
		Inline: true,
	}
}

func renderPopulation(p *message.Printer, loaded meta.Loaded[ps2.WorldsPopulation]) *discordgo.MessageEmbed {
	population := loaded.Value
	worlds := slices.Clone(population.Worlds)
	sort.Slice(worlds, func(i, j int) bool {
		return worlds[i].Id < worlds[j].Id
	})
	fields := make([]*discordgo.MessageEmbedField, len(worlds))
	for i, world := range worlds {
		fields[i] = RenderWorldTotalPopulation(p, world)
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  p.Sprintf("Total population - %d", population.Total),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: p.Sprintf("Source: %s", loaded.Source),
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

func RenderAlertTiming(p *message.Printer, alert ps2.Alert) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name: p.Sprintf("Period"),
		Value: p.Sprintf(
			"%s - %s (Ends %s)",
			renderTime(alert.StartedAt),
			renderTime(alert.StartedAt.Add(alert.Duration)),
			renderRelativeTime(alert.StartedAt.Add(alert.Duration)),
		),
	}
}

func RenderAlertTerritoryControl(p *message.Printer, alert ps2.Alert) *discordgo.MessageEmbedField {
	b := strings.Builder{}
	renderStatPerFactions(p, &b, alert.TerritoryControl)
	return &discordgo.MessageEmbedField{
		Name:  p.Sprintf("Territory Control"),
		Value: b.String(),
	}
}

func RenderWorldAlerts(p *message.Printer, alerts ps2.Alerts) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, len(alerts)*3)
	for _, v := range alerts {
		fields = append(fields, RenderAlertTitle(v), RenderAlertTiming(p, v), RenderAlertTerritoryControl(p, v))
	}
	return fields
}

func renderWorldDetailedAlerts(p *message.Printer, worldName string, loaded meta.Loaded[ps2.Alerts]) *discordgo.MessageEmbed {
	alerts := loaded.Value
	if len(alerts) == 0 {
		return &discordgo.MessageEmbed{
			Title: p.Sprintf("%s - No alerts", worldName),
			Footer: &discordgo.MessageEmbedFooter{
				Text: p.Sprintf("Source: %s", loaded.Source),
			},
			Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
		}
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  p.Sprintf("%s alerts", worldName),
		Fields: RenderWorldAlerts(p, alerts),
		Footer: &discordgo.MessageEmbedFooter{
			Text: p.Sprintf("Source: %s", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}

func renderAlerts(p *message.Printer, loaded meta.Loaded[ps2.Alerts]) []*discordgo.MessageEmbed {
	alerts := loaded.Value
	if len(alerts) == 0 {
		return []*discordgo.MessageEmbed{
			{
				Title: p.Sprintf("No alerts"),
				Footer: &discordgo.MessageEmbedFooter{
					Text: p.Sprintf("Source: %s", loaded.Source),
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
		embeds = append(embeds, renderWorldDetailedAlerts(
			p, worldName, loaded,
		))
	}
	return embeds
}
