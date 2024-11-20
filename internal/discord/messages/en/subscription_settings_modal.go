package en_messages

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/discord"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

var subscriptionModalTitles = map[string]string{
	discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[ps2_platforms.PC]:     "Subscription Settings (PC)",
	discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[ps2_platforms.PS4_EU]: "Subscription Settings (PS4 EU)",
	discord.SUBSCRIPTION_MODAL_CUSTOM_IDS[ps2_platforms.PS4_US]: "Subscription Settings (PS4 US)",
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
						Label:       "Which outfits do you want to track?",
						Placeholder: "Enter the outfit tags separated by comma",
						Style:       discordgo.TextInputShort,
						Value:       strings.Join(outfitTags, ", "),
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "characters",
						Label:       "Which characters do you want to track?",
						Placeholder: "Enter the character names separated by comma",
						Style:       discordgo.TextInputParagraph,
						Value:       strings.Join(characterNames, ", "),
					},
				},
			},
		},
	}, nil
}
