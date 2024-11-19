package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) WorldTerritoryControl(loaded meta.Loaded[ps2.WorldTerritoryControl]) (*discordgo.WebhookEdit, *discord.Error) {
	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			RenderWorldTerritoryControl(loaded),
		},
	}, nil
}
