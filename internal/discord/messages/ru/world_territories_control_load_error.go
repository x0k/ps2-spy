package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) WorldTerritoryControlLoadError(worldId ps2.WorldId, err error) (*discordgo.WebhookEdit, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Ошибка загрузки информации о контроле территорий для" + string(worldId),
		Err: err,
	}
}
