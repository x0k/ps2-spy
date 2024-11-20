package discord_messages

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	"golang.org/x/text/message"
)

func renderZoneTerritoryControl(p *message.Printer, zone ps2.ZoneTerritoryControl) string {
	b := strings.Builder{}
	if zone.IsOpen {
		b.WriteString(p.Sprintf("Unlocked "))
	} else {
		b.WriteString(p.Sprintf("Locked "))
	}
	b.WriteString(renderTime(zone.Since))
	b.WriteString(" (")
	b.WriteString(renderRelativeTime(zone.Since))
	if zone.ControlledBy == factions.None {
		b.WriteByte(')')
	} else {
		b.WriteString(p.Sprintf(") by `"))
		b.WriteString(factions.FactionNameById(zone.ControlledBy))
		b.WriteByte('`')
	}
	if !zone.IsOpen {
		return b.String()
	}
	b.WriteString(p.Sprintf("\nStatus: _"))
	if zone.IsStable {
		b.WriteString(p.Sprintf("Stable"))
	} else {
		b.WriteString(p.Sprintf("Unstable"))
	}
	b.WriteString(p.Sprintf("_\nAlerts: _"))
	if zone.HasAlerts {
		b.WriteString(p.Sprintf("Yes"))
	} else {
		b.WriteString(p.Sprintf("No"))
	}
	b.WriteString("_\n")
	RenderStatPerFactions(p, &b, zone.StatPerFactions)
	return b.String()
}

func RenderZoneTerritoryControl(p *message.Printer, zone ps2.ZoneTerritoryControl) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Name:  ps2.ZoneNameById(zone.Id),
		Value: renderZoneTerritoryControl(p, zone),
	}
}

func RenderZoneTerritoryControls(p *message.Printer, zones []ps2.ZoneTerritoryControl) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, len(zones))
	for _, zone := range zones {
		fields = append(fields, RenderZoneTerritoryControl(p, zone))
	}
	return fields
}

func renderWorldTerritoryControl(p *message.Printer, loaded meta.Loaded[ps2.WorldTerritoryControl]) *discordgo.MessageEmbed {
	world := loaded.Value
	return &discordgo.MessageEmbed{
		Type:   discordgo.EmbedTypeRich,
		Title:  ps2.WorldNameById(world.Id),
		Fields: RenderZoneTerritoryControls(p, world.Zones),
		Footer: &discordgo.MessageEmbedFooter{
			Text: p.Sprintf("Source: %s", loaded.Source),
		},
		Timestamp: loaded.UpdatedAt.Format(time.RFC3339),
	}
}
