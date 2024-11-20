package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (m *messages) OutfitTagsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) (*discordgo.InteractionResponseData, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Не удалось загрузить теги аутфитов для " + string(platform) + " платформы",
		Err: err,
	}
}
