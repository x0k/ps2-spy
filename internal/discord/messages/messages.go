package discord_messages

import (
	"github.com/x0k/ps2-spy/internal/discord"
	en_messages "github.com/x0k/ps2-spy/internal/discord/messages/en"
	ru_messages "github.com/x0k/ps2-spy/internal/discord/messages/ru"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type localizedMessages struct {
	messages map[discord.Locale]discord.Messages
}

func New() *localizedMessages {
	return &localizedMessages{
		messages: map[discord.Locale]discord.Messages{
			discord.EN: en_messages.New(),
			discord.RU: ru_messages.New(),
		},
	}
}

func (m *localizedMessages) About() discord.Localized {
	return func(locale discord.Locale) string {
		return m.messages[locale].About()
	}
}

func (m *localizedMessages) CharacterLogin(char ps2.Character) discord.LocalizedMessage {
	return func(locale discord.Locale) (string, *discord.StringError) {
		return m.messages[locale].CharacterLogin(char)
	}
}

func (m *localizedMessages) CharacterLoadError(characterId ps2.CharacterId, err error) discord.LocalizedMessage {
	return func(locale discord.Locale) (string, *discord.StringError) {
		return m.messages[locale].CharacterLoadError(characterId, err)
	}
}
