package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func (m *messages) SubscriptionSettingsUpdate(entities discord.TrackableEntities[[]string, []string]) (*discordgo.WebhookEdit, *discord.Error) {
	content := RenderSubscriptionsSettingsUpdate(entities)
	return &discordgo.WebhookEdit{
		Content: &content,
	}, nil
}
