package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
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
	WorldAlertsLoadError(string, ps2.WorldId, error) (*discordgo.WebhookEdit, *Error)
	WorldAlerts(worldName string, alerts meta.Loaded[ps2.Alerts]) (*discordgo.WebhookEdit, *Error)
	GlobalAlertsLoadError(provider string, err error) (*discordgo.WebhookEdit, *Error)
	GlobalAlerts(alerts meta.Loaded[ps2.Alerts]) (*discordgo.WebhookEdit, *Error)
	OnlineMembersLoadError(channelId ChannelId, platform ps2_platforms.Platform, err error) (*discordgo.WebhookEdit, *Error)
	OutfitsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) (*discordgo.WebhookEdit, *Error)
	MembersOnline(
		outfitCharacters map[ps2.OutfitId][]ps2.Character,
		characters []ps2.Character,
		outfits map[ps2.OutfitId]ps2.Outfit,
	) (*discordgo.WebhookEdit, *Error)

	SubscriptionSettingsLoadError(channelId ChannelId, platform ps2_platforms.Platform, err error) (*discordgo.InteractionResponseData, *Error)
	OutfitTagsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) (*discordgo.InteractionResponseData, *Error)
	CharacterNamesLoadError(characterIds []ps2.CharacterId, platform ps2_platforms.Platform, err error) (*discordgo.InteractionResponseData, *Error)
	SubscriptionSettingsModal(
		customId string,
		outfitTags []string,
		characterNames []string,
	) (*discordgo.InteractionResponseData, *Error)
}

type LocalizedMessage = func(locale Locale) (string, *Error)

type LocalizedEdit = func(locale Locale) (*discordgo.WebhookEdit, *Error)

type LocalizedResponse = func(locale Locale) (*discordgo.InteractionResponseData, *Error)

type LocalizedMessages interface {
	CharacterLogin(ps2.Character) LocalizedMessage
	CharacterLoadError(ps2.CharacterId, error) LocalizedMessage

	About() LocalizedEdit
	InvalidPopulationType(string, error) LocalizedEdit
	GlobalPopulationLoadError(provider string, err error) LocalizedEdit
	WorldPopulationLoadError(provider string, worldId ps2.WorldId, err error) LocalizedEdit
	GlobalPopulation(meta.Loaded[ps2.WorldsPopulation]) LocalizedEdit
	WorldPopulation(meta.Loaded[ps2.DetailedWorldPopulation]) LocalizedEdit
	WorldTerritoryControlLoadError(ps2.WorldId, error) LocalizedEdit
	WorldTerritoryControl(meta.Loaded[ps2.WorldTerritoryControl]) LocalizedEdit
	WorldAlertsLoadError(provider string, worldId ps2.WorldId, err error) LocalizedEdit
	WorldAlerts(worldName string, alerts meta.Loaded[ps2.Alerts]) LocalizedEdit
	GlobalAlertsLoadError(provider string, err error) LocalizedEdit
	GlobalAlerts(alerts meta.Loaded[ps2.Alerts]) LocalizedEdit
	OnlineMembersLoadError(channelId ChannelId, platform ps2_platforms.Platform, err error) LocalizedEdit
	OutfitsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) LocalizedEdit
	MembersOnline(
		outfitCharacters map[ps2.OutfitId][]ps2.Character,
		characters []ps2.Character,
		outfits map[ps2.OutfitId]ps2.Outfit,
	) LocalizedEdit

	SubscriptionSettingsLoadError(channelId ChannelId, platform ps2_platforms.Platform, err error) LocalizedResponse
	OutfitTagsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) LocalizedResponse
	CharacterNamesLoadError(characterIds []ps2.CharacterId, platform ps2_platforms.Platform, err error) LocalizedResponse
	SubscriptionSettingsModal(
		customId string,
		outfitTags []string,
		characterNames []string,
	) LocalizedResponse
}
