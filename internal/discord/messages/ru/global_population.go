package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) GlobalPopulation(loaded meta.Loaded[ps2.WorldsPopulation]) (*discordgo.WebhookEdit, *discord.Error) {
	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			renderPopulation(loaded),
		},
	}, nil
}
