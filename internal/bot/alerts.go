package bot

import (
	"fmt"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-feed/internal/ps2"
)

func renderAlertTitle(alert ps2.Alert) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:  alert.AlertName,
		Value: alert.AlertDescription,
	}
}

func renderAlertTiming(alert ps2.Alert) *discordgo.MessageEmbedField {
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

func renderAlertTerritoryControl(alert ps2.Alert) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:  "Territory Control",
		Value: renderStatsByFactions(alert.TerritoryControl),
	}
}

func renderWorldAlerts(alerts ps2.Alerts) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, len(alerts)*3)
	for _, v := range alerts {
		fields = append(fields, renderAlertTitle(v), renderAlertTiming(v), renderAlertTerritoryControl(v))
	}
	return fields
}

func renderWorldDetailedAlerts(worldName string, alerts ps2.Alerts, alertsSource string, updatedAt time.Time) *discordgo.MessageEmbed {
	if len(alerts) == 0 {
		return &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s - No alerts", worldName),
		}
	}
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  fmt.Sprintf("%s alerts", worldName),
		Fields: renderWorldAlerts(alerts),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Source: %s", alertsSource),
		},
		Timestamp: updatedAt.Format(time.RFC3339),
	}
}

func renderAlerts(alerts ps2.Alerts, alertsSource string, updatedAt time.Time) []*discordgo.MessageEmbed {
	if len(alerts) == 0 {
		return []*discordgo.MessageEmbed{
			{
				Title: "No alerts",
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
		embeds = append(embeds, renderWorldDetailedAlerts(
			worldName, alerts, alertsSource, updatedAt,
		))
	}
	return embeds
}
