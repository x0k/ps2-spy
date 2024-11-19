package ru_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
)

func (m *messages) GlobalPopulationLoadError(provider string, err error) (*discordgo.WebhookEdit, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Ошибка загрузки глобальной популяции " + provider,
		Err: err,
	}
}
