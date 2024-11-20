package ru_messages

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var subscriptionModalTitles = map[string]string{
	discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[ps2_platforms.PC]:     "Настройки подписки (ПК)",
	discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[ps2_platforms.PS4_EU]: "Настройки подписки (PS4 EU)",
	discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[ps2_platforms.PS4_US]: "Настройки подписки (PS4 US)",
}

func (m *messages) SubscriptionSettingsModal(
	customId string,
	outfitTags []string,
	characterNames []string,
) (*discordgo.InteractionResponseData, *discord.Error) {
	return &discordgo.InteractionResponseData{
		CustomID: customId,
		Title:    subscriptionModalTitles[customId],
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "outfits",
						Label:       "Какие аутфиты вы хотите отслеживать?",
						Placeholder: "Введите теги аутфитов через запятую",
						Style:       discordgo.TextInputShort,
						Value:       strings.Join(outfitTags, ", "),
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "characters",
						Label:       "Каких персонажи вы хотите отслеживать?",
						Placeholder: "Введите имена персонажей через запятую",
						Style:       discordgo.TextInputParagraph,
						Value:       strings.Join(characterNames, ", "),
					},
				},
			},
		},
	}, nil
}
