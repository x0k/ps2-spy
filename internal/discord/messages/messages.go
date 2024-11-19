package discord_messages

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	en_messages "github.com/x0k/ps2-spy/internal/discord/messages/en"
	ru_messages "github.com/x0k/ps2-spy/internal/discord/messages/ru"
	"github.com/x0k/ps2-spy/internal/meta"
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

func (m *localizedMessages) CharacterLogin(char ps2.Character) discord.LocalizedMessage {
	return func(locale discord.Locale) (string, *discord.Error) {
		return m.messages[locale].CharacterLogin(char)
	}
}

func (m *localizedMessages) CharacterLoadError(characterId ps2.CharacterId, err error) discord.LocalizedMessage {
	return func(locale discord.Locale) (string, *discord.Error) {
		return m.messages[locale].CharacterLoadError(characterId, err)
	}
}

func (m *localizedMessages) About() discord.LocalizedResponse {
	return func(locale discord.Locale) (*discordgo.WebhookEdit, *discord.Error) {
		return m.messages[locale].About()
	}
}

func (m *localizedMessages) InvalidPopulationType(popType string, err error) discord.LocalizedResponse {
	return func(locale discord.Locale) (*discordgo.WebhookEdit, *discord.Error) {
		return m.messages[locale].InvalidPopulationType(popType, err)
	}
}

func (m *localizedMessages) GlobalPopulationLoadError(provider string, err error) discord.LocalizedResponse {
	return func(locale discord.Locale) (*discordgo.WebhookEdit, *discord.Error) {
		return m.messages[locale].GlobalPopulationLoadError(provider, err)
	}
}

func (m *localizedMessages) WorldPopulationLoadError(provider string, worldId ps2.WorldId, err error) discord.LocalizedResponse {
	return func(locale discord.Locale) (*discordgo.WebhookEdit, *discord.Error) {
		return m.messages[locale].WorldPopulationLoadError(provider, worldId, err)
	}
}

func (m *localizedMessages) GlobalPopulation(population meta.Loaded[ps2.WorldsPopulation]) discord.LocalizedResponse {
	return func(locale discord.Locale) (*discordgo.WebhookEdit, *discord.Error) {
		return m.messages[locale].GlobalPopulation(population)
	}
}

func (m *localizedMessages) WorldPopulation(population meta.Loaded[ps2.DetailedWorldPopulation]) discord.LocalizedResponse {
	return func(locale discord.Locale) (*discordgo.WebhookEdit, *discord.Error) {
		return m.messages[locale].WorldPopulation(population)
	}
}
