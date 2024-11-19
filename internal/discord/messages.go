package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type Error struct {
	Msg string
	Err error
}

type Messages interface {
	CharacterLogin(ps2.Character) (string, *Error)
	CharacterLoadError(ps2.CharacterId, error) (string, *Error)

	About() (*discordgo.WebhookEdit, *Error)
	InvalidPopulationType(string, error) (*discordgo.WebhookEdit, *Error)
	GlobalPopulation(meta.Loaded[ps2.WorldsPopulation]) (*discordgo.WebhookEdit, *Error)
	GlobalPopulationLoadError(provider string, err error) (*discordgo.WebhookEdit, *Error)
	WorldPopulation(meta.Loaded[ps2.DetailedWorldPopulation]) (*discordgo.WebhookEdit, *Error)
	WorldPopulationLoadError(provider string, worldId ps2.WorldId, err error) (*discordgo.WebhookEdit, *Error)
	WorldTerritoryControlLoadError(ps2.WorldId, error) (*discordgo.WebhookEdit, *Error)
	WorldTerritoryControl(meta.Loaded[ps2.WorldTerritoryControl]) (*discordgo.WebhookEdit, *Error)
}

type LocalizedMessage = func(locale Locale) (string, *Error)

type LocalizedResponse = func(locale Locale) (*discordgo.WebhookEdit, *Error)

type LocalizedMessages interface {
	CharacterLogin(ps2.Character) LocalizedMessage
	CharacterLoadError(ps2.CharacterId, error) LocalizedMessage

	About() LocalizedResponse
	InvalidPopulationType(string, error) LocalizedResponse
	GlobalPopulationLoadError(provider string, err error) LocalizedResponse
	WorldPopulationLoadError(provider string, worldId ps2.WorldId, err error) LocalizedResponse
	GlobalPopulation(meta.Loaded[ps2.WorldsPopulation]) LocalizedResponse
	WorldPopulation(meta.Loaded[ps2.DetailedWorldPopulation]) LocalizedResponse
	WorldTerritoryControlLoadError(ps2.WorldId, error) LocalizedResponse
	WorldTerritoryControl(meta.Loaded[ps2.WorldTerritoryControl]) LocalizedResponse
}
