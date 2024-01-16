package channelsetup

import (
	"context"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/x0k/ps2-spy/internal/bot/handlers"
	"github.com/x0k/ps2-spy/internal/loaders"
)

type SubscriptionSettings struct {
	Outfits    []string
	Characters []string
}

func New(
	loader loaders.KeyedLoader[[2]string, SubscriptionSettings],
) handlers.InteractionHandler {
	return handlers.ShowModal(func(ctx context.Context, log *slog.Logger, s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		platform := i.ApplicationCommandData().Options[0].Name
		settings, err := loader.Load(ctx, [2]string{i.ChannelID, platform})
		if err != nil {
			return nil, err
		}
		return &discordgo.InteractionResponseData{
			CustomID: handlers.CHANNEL_SETUP_MODAL,
			Title:    "Channel Setup",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "outfits",
							Label:       "Which outfits do you want to track?",
							Placeholder: "Enter the outfit tags separated by comma",
							Style:       discordgo.TextInputShort,
							Value:       strings.Join(settings.Outfits, ", "),
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "suggestions",
							Label:       "Which characters do you want to track?",
							Placeholder: "Enter the character names separated by comma",
							Style:       discordgo.TextInputParagraph,
							Value:       strings.Join(settings.Characters, ", "),
						},
					},
				},
			},
		}, nil
	})
}
