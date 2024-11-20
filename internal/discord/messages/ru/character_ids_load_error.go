package ru_messages

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

func (m *messages) CharacterIdsLoadError(characterNames []string, platform ps2_platforms.Platform, err error) (*discordgo.WebhookEdit, *discord.Error) {
	return nil, &discord.Error{
		Msg: "Ошибка загрузки идентификаторов персонажей " + strings.Join(characterNames, ", ") + " (" + string(platform) + ")",
		Err: err,
	}
}
