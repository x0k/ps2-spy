package channelsetup

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
)

func New() handlers.InteractionHandler {
	return handlers.ShowModal(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		return &discordgo.InteractionResponseData{
			CustomID: "modals_survey_" + i.Interaction.Member.User.ID,
			Title:    "Channel Setup",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "opinion",
							Label:       "What is your opinion on them?",
							Style:       discordgo.TextInputShort,
							Placeholder: "Don't be shy, share your opinion with us",
							Required:    true,
							MaxLength:   300,
							MinLength:   10,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:  "suggestions",
							Label:     "What would you suggest to improve them?",
							Style:     discordgo.TextInputParagraph,
							Required:  false,
							MaxLength: 2000,
						},
					},
				},
			},
		}, nil
	})
}
