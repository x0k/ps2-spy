package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) MembersOnline(
	outfitCharacters map[ps2.OutfitId][]ps2.Character,
	characters []ps2.Character,
	outfits map[ps2.OutfitId]ps2.Outfit,
) (*discordgo.WebhookEdit, *discord.Error) {
	content := RenderOnline(outfitCharacters, characters, outfits)
	return &discordgo.WebhookEdit{
		Content: &content,
	}, nil
}
