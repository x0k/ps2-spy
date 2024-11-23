package discord_messages

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/meta"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Messages struct{}

func New() *Messages {
	return &Messages{}
}

func (m *Messages) CharacterLogin(char ps2.Character) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		if char.OutfitTag != "" {
			return p.Sprintf(
				"[%s] %s (%s) is now online (%s)",
				char.OutfitTag,
				char.Name,
				ps2_factions.FactionNameById(char.FactionId),
				ps2.WorldNameById(char.WorldId),
			), nil
		}
		return p.Sprintf(
			"%s (%s) is now online (%s)",
			char.Name,
			ps2_factions.FactionNameById(char.FactionId),
			ps2.WorldNameById(char.WorldId),
		), nil
	}
}

func (m *Messages) CharacterLogout(char ps2.Character) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		if char.OutfitTag != "" {
			return p.Sprintf(
				"[%s] %s (%s) is now offline (%s)",
				char.OutfitTag,
				char.Name,
				ps2_factions.FactionNameById(char.FactionId),
				ps2.WorldNameById(char.WorldId),
			), nil
		}
		return p.Sprintf(
			"%s (%s) is now offline (%s)",
			char.Name,
			ps2_factions.FactionNameById(char.FactionId),
			ps2.WorldNameById(char.WorldId),
		), nil
	}
}

func (m *Messages) OutfitLoadError(outfitId ps2.OutfitId, platform ps2_platforms.Platform, err error) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		return "", &discord.Error{
			Msg: p.Sprintf("Failed to load outfit: %s (%s)", outfitId, platform),
			Err: err,
		}
	}
}

func (m *Messages) CharactersLoadError(characterIds []ps2.CharacterId, platform ps2_platforms.Platform, err error) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		return "", &discord.Error{
			Msg: p.Sprintf("Failed to load characters: %v (%s)", characterIds, platform),
			Err: err,
		}
	}
}

func (m *Messages) OutfitMembersUpdate(outfit ps2.Outfit, change diff.Diff[ps2.Character]) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		builder := strings.Builder{}
		builder.WriteString(p.Sprintf("Update of "))
		builder.WriteString(outfit.Name)
		builder.WriteString(" [")
		builder.WriteString(outfit.Tag)
		builder.WriteString(p.Sprintf("] outfit members:"))
		if len(change.ToAdd) > 0 {
			builder.WriteString(p.Sprintf("\n**Welcome to the outfit:**"))
			for i := range change.ToAdd {
				builder.WriteString("\n- ")
				builder.WriteString(change.ToAdd[i].Name)
			}
		}
		if len(change.ToDel) > 0 {
			builder.WriteString(p.Sprintf("\n**Left the outfit:**"))
			for i := range change.ToDel {
				builder.WriteString("\n- ")
				builder.WriteString(change.ToDel[i].Name)
			}
		}
		return builder.String(), nil
	}
}

func (m *Messages) CharacterLoadError(characterId ps2.CharacterId, err error) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		return "", &discord.Error{
			Msg: p.Sprintf("Failed to load character: %s", characterId),
			Err: err,
		}
	}
}

func (m *Messages) FacilityLoadError(facilityId ps2.FacilityId, err error) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		return "", &discord.Error{
			Msg: p.Sprintf("Failed to load facility: %s", facilityId),
			Err: err,
		}
	}
}

func (m *Messages) FacilityControl(
	worldId ps2.WorldId,
	outfit ps2.Outfit,
	facility ps2.Facility,
) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		// TODO: Fix this
		// outfit[tag] захватил Regent Rock Garrison (Large Outpost) в Indar (server)
		return p.Sprintf(
			"%s [%s] captured %s (%s) on %s (%s)",
			outfit.Name,
			outfit.Tag,
			facility.Name,
			facility.Type,
			ps2.ZoneNameById(facility.ZoneId),
			ps2.WorldNameById(worldId),
		), nil
	}
}

func (m *Messages) FacilityLoss(
	worldId ps2.WorldId,
	outfit ps2.Outfit,
	facility ps2.Facility,
) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		return p.Sprintf(
			"%s [%s] lost %s (%s) on %s (%s)",
			outfit.Name,
			outfit.Tag,
			facility.Name,
			facility.Type,
			ps2.ZoneNameById(facility.ZoneId),
			ps2.WorldNameById(worldId),
		), nil
	}
}

func (m *Messages) About() discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		content := p.Sprintf(`# PlanetSide 2 Spy

Simple discord bot for PlanetSide 2 outfits

## Links

- [GitHub](https://github.com/x0k/ps2-spy)
		
`)
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	}
}

func (m *Messages) InvalidPopulationType(popType string, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Invalid population type: %s", popType),
			Err: err,
		}
	}
}

func (m *Messages) GlobalPopulationLoadError(provider string, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load global population with %s", provider),
			Err: err,
		}
	}
}

func (m *Messages) WorldPopulationLoadError(provider string, worldId ps2.WorldId, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load %s population with %s", ps2.WorldNameById(worldId), provider),
			Err: err,
		}
	}
}

func (m *Messages) GlobalPopulation(population meta.Loaded[ps2.WorldsPopulation]) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				renderPopulation(p, population),
			},
		}, nil
	}
}

func (m *Messages) WorldPopulation(population meta.Loaded[ps2.DetailedWorldPopulation]) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				renderWorldDetailedPopulation(p, population),
			},
		}, nil
	}
}

func (m *Messages) WorldTerritoryControlLoadError(worldId ps2.WorldId, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load %s territory control", ps2.WorldNameById(worldId)),
			Err: err,
		}
	}
}

func (m *Messages) WorldTerritoryControl(control meta.Loaded[ps2.WorldTerritoryControl]) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				renderWorldTerritoryControl(p, control),
			},
		}, nil
	}
}

func (m *Messages) WorldAlertsLoadError(provider string, worldId ps2.WorldId, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load world alerts for %s from %s", ps2.WorldNameById(worldId), provider),
			Err: err,
		}
	}
}

func (m *Messages) WorldAlerts(worldName string, worldAlerts meta.Loaded[ps2.Alerts]) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				renderWorldDetailedAlerts(p, worldName, worldAlerts),
			},
		}, nil
	}
}

func (m *Messages) GlobalAlertsLoadError(provider string, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load global alerts from %s", provider),
			Err: err,
		}
	}
}

func (m *Messages) GlobalAlerts(alerts meta.Loaded[ps2.Alerts]) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		embeds := renderAlerts(p, alerts)
		return &discordgo.WebhookEdit{
			Embeds: &embeds,
		}, nil
	}
}

func (m *Messages) OnlineMembersLoadError(channelId discord.ChannelId, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load online members for %s channel (%s)", channelId, platform),
			Err: err,
		}
	}
}

func (m *Messages) OutfitsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load outfits %v (%s)", outfitIds, platform),
			Err: err,
		}
	}
}

func (m *Messages) MembersOnline(
	outfitCharacters map[ps2.OutfitId][]ps2.Character,
	characters []ps2.Character,
	outfits map[ps2.OutfitId]ps2.Outfit,
) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		content := renderOnline(p, outfitCharacters, characters, outfits)
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	}
}

func (m *Messages) OutfitIdsLoadError(outfitTags []string, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load outfits by tags %v (%s)", outfitTags, platform),
			Err: err,
		}
	}
}

func (m *Messages) CharacterIdsLoadError(characterNames []string, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load characters %v (%s)", characterNames, platform),
			Err: err,
		}
	}
}

func (m *Messages) TrackingSettingsSaveError(channelId discord.ChannelId, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to save tracking settings for %s channel (%s)", channelId, platform),
			Err: err,
		}
	}
}

func (m *Messages) TrackingSettingsOutfitTagsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Settings are saved, but failed to load outfit tags %v (%s)", outfitIds, platform),
			Err: err,
		}
	}
}

func (m *Messages) TrackingSettingsCharacterNamesLoadError(characterIds []ps2.CharacterId, platform ps2_platforms.Platform, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Settings are saved, but failed to load character names %v (%s)", characterIds, platform),
			Err: err,
		}
	}
}

func (m *Messages) TrackingSettingsUpdate(entities discord.TrackableEntities[[]string, []string]) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		content := renderSubscriptionsSettingsUpdate(p, entities)
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	}
}

func (m *Messages) TrackingSettingsLoadError(channelId discord.ChannelId, platform ps2_platforms.Platform, err error) discord.Response {
	return func(p *message.Printer) (*discordgo.InteractionResponseData, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load tracking settings for %s channel (%s)", channelId, platform),
			Err: err,
		}
	}
}

func (m *Messages) OutfitTagsLoadError(outfitIds []ps2.OutfitId, platform ps2_platforms.Platform, err error) discord.Response {
	return func(p *message.Printer) (*discordgo.InteractionResponseData, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load outfit tags for %v (%s)", outfitIds, platform),
			Err: err,
		}
	}
}

func (m *Messages) CharacterNamesLoadError(characterIds []ps2.CharacterId, platform ps2_platforms.Platform, err error) discord.Response {
	return func(p *message.Printer) (*discordgo.InteractionResponseData, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to load character names for %v (%s)", characterIds, platform),
			Err: err,
		}
	}
}

func (m *Messages) TrackingSettingsModal(
	customId string,
	outfitTags []string,
	characterNames []string,
) discord.Response {
	return func(p *message.Printer) (*discordgo.InteractionResponseData, *discord.Error) {
		var trackingModalTitles = map[string]string{
			discord.TRACKING_MODAL_CUSTOM_IDS[ps2_platforms.PC]:     p.Sprintf("Tracking Settings (PC)"),
			discord.TRACKING_MODAL_CUSTOM_IDS[ps2_platforms.PS4_EU]: p.Sprintf("Tracking Settings (PS4 EU)"),
			discord.TRACKING_MODAL_CUSTOM_IDS[ps2_platforms.PS4_US]: p.Sprintf("Tracking Settings (PS4 US)"),
		}
		return &discordgo.InteractionResponseData{
			CustomID: customId,
			Title:    trackingModalTitles[customId],
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "outfits",
							Label:       p.Sprintf("Which outfits do you want to track?"),
							Placeholder: p.Sprintf("Enter the outfit tags separated by comma"),
							Style:       discordgo.TextInputShort,
							Value:       strings.Join(outfitTags, ", "),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "characters",
							Label:       p.Sprintf("Which characters do you want to track?"),
							Placeholder: p.Sprintf("Enter the character names separated by comma"),
							Style:       discordgo.TextInputParagraph,
							Value:       strings.Join(characterNames, ", "),
						},
					},
				},
			},
		}, nil
	}
}

func (m *Messages) ChannelLanguageParseError(channelId discord.ChannelId, lang string, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to parse language %q", lang),
			Err: err,
		}
	}
}

func (m *Messages) ChannelLanguageSaveError(channelId discord.ChannelId, lang language.Tag, err error) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		return nil, &discord.Error{
			Msg: p.Sprintf("Failed to save language %q", lang),
			Err: err,
		}
	}
}

func (m *Messages) ChannelLanguageSaved(channelId discord.ChannelId, lang language.Tag) discord.Edit {
	return func(p *message.Printer) (*discordgo.WebhookEdit, *discord.Error) {
		content := p.Sprintf("Language for this channel has been set to %q", lang.String())
		return &discordgo.WebhookEdit{
			Content: &content,
		}, nil
	}
}

func (m *Messages) OnlineCountTitleUpdate(title string, count int) discord.Message {
	return func(p *message.Printer) (string, *discord.Error) {
		onlineCount := p.Sprintf("%d・online", count)
		const separator = "│"
		index := strings.LastIndex(title, separator)
		if index == -1 {
			if count == 0 {
				return title, nil
			}
			return title + separator + onlineCount, nil
		} else {
			originalTitle := string([]rune(title)[:index])
			if count == 0 {
				return originalTitle, nil
			}
			return originalTitle + separator + onlineCount, nil
		}
	}
}
