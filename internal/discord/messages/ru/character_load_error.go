package ru_messages

import (
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/ps2"
)

func (m *messages) CharacterLoadError(characterId ps2.CharacterId, err error) (string, *discord.StringError) {
	return "", &discord.StringError{
		Msg: "Не удалось загрузить персонажа: " + string(characterId),
		Err: err,
	}
}
